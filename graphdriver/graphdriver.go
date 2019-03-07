package graphdriver

import (
    // "github.com/docker/docker/pkg/archive"
    // "github.com/docker/docker/pkg/containerfs"
    // "log"
    "fmt"
    "os"
    "path"
    "context"
    "io"
    "time"
    // "strconv"

    "github.com/docker/docker/pkg/idtools"
    "github.com/docker/docker/pkg/containerfs"
    "github.com/docker/docker/pkg/locker"
    "golang.org/x/sys/unix"
    "github.com/docker/docker/pkg/mount"
    "github.com/docker/docker/pkg/archive"
    "github.com/docker/docker/pkg/chrootarchive"
    // rsystem "github.com/opencontainers/runc/libcontainer/system"
    "github.com/docker/docker/pkg/directory"
    "github.com/docker/docker/daemon/graphdriver"
    graphPlugin "github.com/docker/go-plugins-helpers/graphdriver"
)

var (
    // untar defines the untar method
    untar = chrootarchive.UntarUncompressed

    // ApplyUncompressedLayer defines the unpack method used by the graph
    // driver
    ApplyUncompressedLayer = chrootarchive.ApplyUncompressedLayer
)

var (
    backingFs             = "<unknown>"
    projectQuotaSupported = false

    indexOff string
)

type Driver struct {
    home          string
    uidMaps       []idtools.IDMap
    gidMaps       []idtools.IDMap
    options       []string
    naiveDiff     graphdriver.DiffDriver
    locker        *locker.Locker
}

func (d *Driver) Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error{
    fmt.Println("\nInit func parameters: ")
    fmt.Printf("  home: %s\n", home)
    fmt.Printf("  options: %s\n", options)
    fmt.Println("  uidMaps: ", uidMaps)
    fmt.Println("  gidMaps: ", gidMaps)
    
    if home == "" {
        fmt.Println("Init func: home is not set, I will set it to default: ~/.mydriver/")
    }
    d.home = home
    d.uidMaps = uidMaps
    d.gidMaps = gidMaps
    d.options = options

    // d.naiveDiff = graphdriver.NewNaiveDiffDriver(d, uidMaps, gidMaps)

    return nil
}

// Create is used to create an image directory, in my design there is only two directories in an image dir.
// In my design, an image is an independent unit, has no parents
// Create(id, parent string, opts *CreateOpts) error
func (d *Driver) Create(id, parent, mountlabel string, storageOpt map[string]string) error{
    fmt.Println("\nCreate func parameters: ")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)
    fmt.Printf("  mountLabel: %s\n", mountlabel)
    fmt.Printf("  storageOpt: %s\n", storageOpt)

    // I also don't care about mountlabel and storageOpt

    return nil
}

// create a readwrite filesystem layer
func (d *Driver) CreateReadWrite(id, parent, mountlabel string, storageOpt map[string]string) error{
    fmt.Println("\nCreateReadWrite func parameters: ")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)
    fmt.Printf("  mountLabel: %s\n", mountlabel)
    fmt.Printf("  storageOpt: %s\n", storageOpt)

    return nil
}

// Remove cleans the directories that are created for this id
func (d *Driver) Remove(id string) error{
    fmt.Printf("\nRemove func parameters: \n")
    fmt.Printf("  id: %s\n", id)

    // If the id is nil, refuse to remove the dir
    if id == "" {
        return fmt.Errorf("refusing to remove the directories: id is empty")
    }

    d.locker.Lock(id)
    defer d.locker.Unlock(id)

    err := os.RemoveAll(path.Join(d.home, id))
    if err != nil {
        return fmt.Errorf("failing in removing directories: encounter errors")
    }

    return nil
}

// Get creates and mounts the required file system for the given id and returns the mount path
func (d *Driver) Get(id, mountLabel string) (containerfs.ContainerFS, error){
    fmt.Printf("\nGet func parameters: \n")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  mountlabel: %s\n", mountLabel)

    d.locker.Lock(id)
    defer d.locker.Unlock(id)

    dir := path.Join(d.home, id)
    if _, err := os.Stat(dir); err != nil {
        return nil, err
    }

    rootUID, rootGID, err := idtools.GetRootUIDGID(d.uidMaps, d.gidMaps)
    if err != nil {
        return nil, err
    }

    mergedDir := path.Join(d.home, "merged")
    
    upperDir := path.Join(d.home, "container_data")

    lowerDir := path.Join(d.home, "image_data")

    workDir := path.Join(d.home, "work")

    if err := idtools.MkdirAndChown(mergedDir, 0700, idtools.Identity{UID: rootUID, GID: rootGID}); err != nil {
        return nil, err
    }

    mountData := "lowerdir=" + lowerDir + ",upperdir=" + upperDir + ",workdir=" + workDir

    mount := func(source string, target string, mType string, flags uintptr, label string) error {
        return nil
    }

    err = mount("overlay", mergedDir, "overlay", 0, mountData)

    if err != nil {
        return nil, fmt.Errorf("failing in mount overlayfs: encounter errors")
    }

    // chown "workdir/work" to the remapped root UID/GID. Overlay fs inside a
    // user namespace requires this to move a directory from lower to upper.
    if err := os.Chown(path.Join(workDir, "work"), rootUID, rootGID); err != nil {
        return nil, err
    }

    return containerfs.NewLocalContainerFS(mergedDir), nil
}

// Put unmounts the mount path created for the give id.
// It also removes the 'merged' directory to force the kernel to unmount the
// overlay mount in other namespaces.
func (d *Driver) Put(id string) error{
    fmt.Printf("\nPut func parameters: \n")
    fmt.Printf("  id: %s\n", id)


    d.locker.Lock(id)
    defer d.locker.Unlock(id)

    dir := path.Join(d.home, id)

    mergedDir := path.Join(dir, "merged")

    if err := unix.Unmount(mergedDir, 0x2); err != nil {
        fmt.Printf("Failed to unmount %s overlay: %s - %v", id, mergedDir, err)
    }

    // Remove the mountpoint here. Removing the mountpoint (in newer kernels)
    // will cause all other instances of this mount in other mount namespaces
    // to be unmounted. This is necessary to avoid cases where an overlay mount
    // that is present in another namespace will cause subsequent mounts
    // operations to fail with ebusy.  We ignore any errors here because this may
    // fail on older kernels which don't have
    // torvalds/linux@8ed936b5671bfb33d89bc60bdcc7cf0470ba52fe applied.
    if err := unix.Rmdir(mergedDir); err != nil && !os.IsNotExist(err) {
        fmt.Printf("Failed to remove %s overlay: %v", id, err)
    }
    return nil
}

// Exists checks to see if the id is already mounted
func (d *Driver) Exists(id string) bool{
    fmt.Printf("\nExists func parameters: \n")
    fmt.Printf("  id: %s\n", id)


    _, err := os.Stat(path.Join(d.home, id))
    return err == nil
}

// Status returns current driver information in a two dimensional string array.
// Output contains "Backing Filesystem" used in this implementation
func (d *Driver) Status() [][2]string{
    fmt.Printf("\nStatus func parameters: \n")

    return [][2]string{
        {"Backing Filesystem", backingFs},
        // {"Supports d_type", strconv.FormatBool(d.supportsDType)},
        // {"Native Overlay Diff", strconv.FormatBool(!useNaiveDiff(d.home))},
    }
}

// GetMetadata returns metadata about the overlay driver such as the LowerDir,
// UpperDir, WorkDir, and MergeDir used to store data
func (d *Driver) GetMetadata(id string) (map[string]string, error){
    fmt.Printf("\nGetMetadata func parameters: \n")
    fmt.Printf("  id: %s\n", id)


    dir := path.Join(d.home, id)
    if _, err := os.Stat(dir); err != nil {
        return nil, err
    }

    metadata := map[string]string{
        "WorkDir":   path.Join(dir, "work"),
        "MergedDir": path.Join(dir, "merged"),
        "UpperDir":  path.Join(dir, "container_data"),
        "LowerDir": path.Join(dir, "image_data"), 
    }

    return metadata, nil
}

// Cleanup any state created by overlay which should be cleaned when daemon
// is being shutdown. For now, we just have to unmount the bind mounted
// we had created
func (d *Driver) Cleanup() error{
    fmt.Printf("\nCleanup func parameters: \n")

    return mount.Unmount(d.home)
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be ""
func (d *Driver) Diff(id, parent string) io.ReadCloser {
    fmt.Printf("\nDiff func parameters: \n")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)
    // if parent is not id's parent 
    // if NaiveDiff can be used
    // if useNaiveDiff(d.home) || !d.isParent(id, parent) {
    //  return d.naiveDiff.Diff(id, parent)
    // }


    dir := path.Join(d.home, id)

    diffPath := path.Join(dir, "container_data")

    fmt.Printf("  Tar with options on %s", diffPath)
    result, err :=  archive.TarWithOptions(diffPath, &archive.TarOptions{
        Compression:    archive.Uncompressed,
        UIDMaps:        d.uidMaps,
        GIDMaps:        d.gidMaps,
        WhiteoutFormat: archive.OverlayWhiteoutFormat,
    })

    if err != nil {
        fmt.Println(err)
    }

    return result
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes
func (d *Driver) Changes(id, parent string) ([]graphPlugin.Change, error){
    fmt.Printf("\nChanges func parameters: \n")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)

    // layerFs := path.Join(d.home, id)

    if parent != "" {
        fmt.Printf("  Failing! parent is not \"\"")
    }
    // parentFs := parent
    
    return []graphPlugin.Change{}, nil
}

// ApplyDiff applies the new layer into a root
// docker pull will call this func
// TODO: this func has bug
func (d *Driver) ApplyDiff(id, parent string, diff io.Reader) (int64, error){
    fmt.Printf("\nApplyDiff func parameters: \n")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)

    options := &archive.TarOptions{UIDMaps: d.uidMaps,
        GIDMaps: d.gidMaps}
    
    layerFs := path.Join(d.home, id)

    start := time.Now().UTC()
    fmt.Println("Start untar layer")
    size, err := ApplyUncompressedLayer(layerFs, diff, options); 

    if err != nil {
        return size, err
    }

    fmt.Printf("Untar time: %vs\n", time.Now().UTC().Sub(start).Seconds())


    dir := path.Join(d.home, id)

    applyDir := path.Join(dir, "image_data")

    fmt.Printf("  Applying tar in %s\n", applyDir)

    return size, nil

    // if err := untar(diff, applyDir, &archive.TarOptions{
    //  UIDMaps:        d.uidMaps,
    //  GIDMaps:        d.gidMaps,
    //  WhiteoutFormat: archive.OverlayWhiteoutFormat,
    // }); err != nil {
    //  return 0, err
    // }

    // return directory.Size(context.TODO(), applyDir)
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory
func (d *Driver) DiffSize(id, parent string) (int64, error){
    fmt.Printf("\nDiffSizee func parameters: \n")
    fmt.Printf("  id: %s\n", id)
    fmt.Printf("  parent: %s\n", parent)

    dir := path.Join(d.home, id)

    diffPath :=  path.Join(dir, "container_data")

    return directory.Size(context.TODO(), diffPath)
}


func (d *Driver) Capabilities() graphdriver.Capabilities{
    fmt.Printf("\nCapabilities func parameters: \n")
    
    return graphdriver.Capabilities{
        ReproducesExactDiffs: false, 
    }
}

// To implement protoDriver
func (d *Driver) String() string {
    return "This is my Driver"
}