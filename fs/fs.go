package fs

import (
	"os"
	// "io"
	"fmt"
	"errors"
	// "reflect"
	"syscall"
	"net/url"
	"net/http"
	"os/signal"
	"io/ioutil"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fuseFS "bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.WithField("fs", "gearFS")
)

type GearFS struct {
	MountPoint string
	IndexImagePath string
	PrivateCachePath string
}

func (g * GearFS) Start() {
	// 1. 检测index image目录、 private cache目录和挂载点目录是否合法
	indexImagePath, err := ValidatePath(g.IndexImagePath)
	if err != nil {
		logrus.Fatalf("indexImagePath: %s is not valid...", g.IndexImagePath)
	}
	privateCachePath, err := ValidatePath(g.PrivateCachePath)
	if err != nil {
		logrus.Fatalf("privateCachePath: %s is not valid...", g.PrivateCachePath)
	}
	mountPoint, err := ValidatePath(g.MountPoint)
	if err != nil {
		logrus.Fatalf("mountPoint: %s is not valid...", g.MountPoint)
	}

	// 2. 在挂载点创建fuse连接
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	// 3. 捕获异常退出，并将mount资源卸载
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func(c *fuse.Conn) {
		<- sigs
		fmt.Println("umounting fuse...")
		for {
			err := c.Close()
			if err != nil {
				fmt.Println(err)
			}
			if err == nil {
				break
			}
		}
		os.Exit(0)
	}(c)

	// 4. 初始化fuse文件系统
	filesys := Init(indexImagePath, privateCachePath)

	// 5. 使用fuse文件系统服务挂载点的fuse连接
	if err := fuseFS.Serve(c, filesys); err != nil {
		fmt.Println(err)
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		fmt.Println(err)
	}
}

func (g * GearFS) StartAndNotify(notify chan int) {
	// 1. 检测index image目录、 private cache目录和挂载点目录是否合法
	indexImagePath, err := ValidatePath(g.IndexImagePath)
	if err != nil {
		logrus.Fatalf("indexImagePath: %s is not valid...", g.IndexImagePath)
	}
	privateCachePath, err := ValidatePath(g.PrivateCachePath)
	if err != nil {
		logrus.Fatalf("privateCachePath: %s is not valid...", g.PrivateCachePath)
	}
	mountPoint, err := ValidatePath(g.MountPoint)
	if err != nil {
		logrus.Fatalf("mountPoint: %s is not valid...", g.MountPoint)
	}

	// 2. 在挂载点创建fuse连接
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	// 3. 捕获异常退出，并将mount资源卸载
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func(c *fuse.Conn) {
		<- sigs
		fmt.Println("umounting fuse...")
		for {
			err := c.Close()
			if err != nil {
				fmt.Println(err)
			}
			if err == nil {
				break
			}
		}
		os.Exit(0)
	}(c)

	// 4. 初始化fuse文件系统
	filesys := Init(indexImagePath, privateCachePath)

	// 5. 使用fuse文件系统服务挂载点的fuse连接
	notify <- 1
	if err := fuseFS.Serve(c, filesys); err != nil {
		fmt.Println(err)
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		fmt.Println(err)
	}
}

func Init(indexImagePath, privateCachePath string) *FS {
	return &FS{
		IndexImagePath: indexImagePath,
		PrivateCachePath: privateCachePath, 
	}
}

type FS struct {
	IndexImagePath string
	PrivateCachePath string
}

func (f *FS) Root() (fs.Node, error) {
	fmt.Println("Root")
	n := &Dir {
		// isRoot: true, 
		dirPath: f.IndexImagePath, 
		privateCachePath: f.PrivateCachePath, 
	}

	return n, nil
}

type Dir struct {
	// isRoot bool
	// rootPath string

	// isDir bool
	dirPath string
	privateCachePath string
}

// TODO: 实际获取每个目录的属性
func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	fmt.Println("dir.Attr()")
	attr.Mode = os.ModeDir | 0755

	return nil
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	fmt.Println("dir.Lookup()")
	path := req.Name

	path = d.dirPath + path

	fInfo, err := os.Lstat(path)
	if err != nil {
		logger.Warnf("Fail to Lstat path %s : %v", path, err)
		return nil, fuse.ENOENT
	}

	if fInfo.IsDir() {
		child := &Dir {
			// isDir: true, 
			dirPath: path+"/", 
			privateCachePath: d.privateCachePath, 
		}
		return child, nil
	} else {
		if fInfo.Mode().IsRegular() {
			child := &File {
				isRegular: true, 
				filePath: path, 
				privateCachePath: d.privateCachePath, 
			}
			return child, nil
		} else {
			child := &File {
				isRegular: false, 
				filePath: path, 
				privateCachePath: d.privateCachePath, 
			}
			return child, nil
		}
	}

	return nil, fuse.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fmt.Println("dir.ReadDirAll()")
	var res []fuse.Dirent
	var path string

	path = d.dirPath

	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Warnf("Fail to read file under dir: %v", err)
	}

	for _, file := range files {
		var de fuse.Dirent
		de.Name = file.Name()
		switch file.Mode() {
		case os.ModeDir: de.Type = fuse.DT_Dir
		case os.ModeSymlink: de.Type = fuse.DT_Link
		case os.ModeNamedPipe: de.Type = fuse.DT_FIFO
		case os.ModeSocket: de.Type = fuse.DT_Socket
		case os.ModeDevice: de.Type = fuse.DT_Block
		case os.ModeCharDevice: de.Type = fuse.DT_Char
		// case os.ModeIrregular: de.Type = DT_Unknown
		default: de.Type = fuse.DT_File
		}
		res = append(res, de)
	}

	return res, nil
}

// used for mkdir
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	fmt.Println("d.Mkdir")
	fmt.Println(req)
	path := req.Name

	path = d.dirPath + path

	mode := req.Mode
	umask := req.Umask

	_, err := os.Lstat(path)
	if err != nil {
		err := os.Mkdir(path, (umask^mode))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Fail to create directory '%s'", path))
		}
		return &Dir {
			dirPath: path+"/", 
			privateCachePath: d.privateCachePath, 
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("cannot create directory '%s': File exists", path))
}

// used for rmdir/rm
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	fmt.Println("d.Remove")
	path := req.Name

	path = d.dirPath + path

	err := os.Remove(path)
	
	return err
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	fmt.Println("dir.Create")
	path := req.Name
	path = d.dirPath + path

	_, err := os.Create(path)
	if err != nil {
		return nil, nil, nil
	}

	child := &File {
		isRegular: true, 
		filePath: path, 
		privateCachePath: d.privateCachePath, 
	}

	return child, child, nil
}

type File struct {
	isRegular bool
	filePath string

	privateCahceName string
	privateCachePath string
}

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	fmt.Println("f.Attr()")
	fmt.Println(f)

	if f.isRegular {
		// 获取文件的cid
		name, err := ioutil.ReadFile(f.filePath)
		if err != nil {
			logger.Warnf("Fail to read filename")
		}
		f.privateCahceName = string(name)

		// 检测private cache中是否存在该文件
		_, err = os.Lstat(filepath.Join(f.privateCachePath, f.privateCahceName))
		if err != nil {
			_, err := http.PostForm("http://localhost:2020"+"/get/"+f.privateCahceName, url.Values{"PATH":{f.privateCachePath}, "PERM":{"0777"}})
			if err != nil {
				logger.Warnf("Fail to get file for %v", err)
			}
			// 修改文件的权限
			IndexFileInfo, err := os.Lstat(f.filePath)
			if err != nil {
				logger.Warnf("Fail to get index file info for %v", err)
			}
			err = os.Chmod(filepath.Join(f.privateCachePath, f.privateCahceName), IndexFileInfo.Mode())
			if err != nil {
				logger.Warnf("Fail to chmod for %v", err)
			}
			err = os.Chown(filepath.Join(f.privateCachePath, f.privateCahceName), int(IndexFileInfo.Sys().(*syscall.Stat_t).Uid), int(IndexFileInfo.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				logger.Warnf("Fail to chown for %v", err)
			}
		}

		fInfo, err := os.Lstat(filepath.Join(f.privateCachePath, f.privateCahceName))
		if err != nil {
			logger.Warnf("Fail to lstat file for %v", err)
		}

		attr.Size = uint64(fInfo.Size())

		IndexFileInfo, err := os.Lstat(f.filePath)
		if err != nil {
			logger.Warnf("Fail to get index file info for %v", err)
		}

		attr.Mode = IndexFileInfo.Mode()
		attr.Mtime = IndexFileInfo.ModTime()
		attr.Uid = IndexFileInfo.Sys().(*syscall.Stat_t).Uid
		attr.Gid = IndexFileInfo.Sys().(*syscall.Stat_t).Gid
	} else {
		IndexFileInfo, err := os.Lstat(f.filePath)
		if err != nil {
			logger.Warnf("Cannot stat file: %v", err)
		}

		attr.Size = uint64(IndexFileInfo.Size())
		attr.Mode = IndexFileInfo.Mode()
		attr.Mtime = IndexFileInfo.ModTime()
		attr.Uid = IndexFileInfo.Sys().(*syscall.Stat_t).Uid
		attr.Gid = IndexFileInfo.Sys().(*syscall.Stat_t).Gid
	}

	fmt.Println(attr)
	return nil
}

func (f *File) Access(ctx context.Context, req *fuse.AccessRequest) error {
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	fmt.Println("f.Open()")
	fmt.Println(f)

	var fileHandler = FileHandler{}

	if f.isRegular {
		// 1. 检查该镜像的私有缓存中是否存在cid文件
		_, err := os.Lstat(filepath.Join(f.privateCachePath, f.privateCahceName))
		if err != nil {
			// 该当前私有缓存中不存在cid文件，向gear client请求将cid文件下载到指定目录
			_, err := http.PostForm("http://localhost:2020"+"/get/"+f.privateCahceName, url.Values{"PATH":{f.privateCachePath}, "PERM":{"0777"}})
			if err != nil {
				logger.Warnf("Fail to get file for %v", err)
			}
			// 修改文件的权限
			IndexFileInfo, err := os.Lstat(f.filePath)
			if err != nil {
				logger.Warnf("Fail to get index file info for %v", err)
			}
			err = os.Chmod(filepath.Join(f.privateCachePath, f.privateCahceName), IndexFileInfo.Mode())
			if err != nil {
				logger.Warnf("Fail to chmod for %v", err)
			}
			err = os.Chown(filepath.Join(f.privateCachePath, f.privateCahceName), int(IndexFileInfo.Sys().(*syscall.Stat_t).Uid), int(IndexFileInfo.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				logger.Warnf("Fail to chown for %v", err)
			}
		}

		// 2. 打开私有缓存中的文件
		file, err := os.Open(filepath.Join(f.privateCachePath, f.privateCahceName))
		if err != nil {
			logger.Warnf("Fail to open file: %v", err)
		}
		fileHandler.f = file

		return &fileHandler, nil
	}

	file, err := os.Open(f.filePath)
	if err != nil {
		logger.Warnf("Fail to open file: %v", err)
	}
	fileHandler.f = file
	
	fmt.Println(fileHandler)
	return &fileHandler, nil
}

func (f *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	fmt.Println("f.Readlink()")
	fmt.Println(f)

	target, err := os.Readlink(f.filePath)
	if err != nil {
		logger.Warnf("Fail to read link for %v", err)
	}

	return target, err
}

// func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {

// }

type FileHandler struct {
	f *os.File
}

func (fh *FileHandler) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	fmt.Println("fh.Release()")
	fmt.Println(fh)

	return fh.f.Close()
}

func (fh *FileHandler) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fmt.Println("fh.Read()")
	fmt.Println(fh)

	buf := make([]byte, req.Size)
	n, err := fh.f.Read(buf)
	resp.Data = buf[:n]

	return err
}

func (fh *FileHandler) ReadAll(ctx context.Context) ([]byte, error) {
	fmt.Println("fh.ReadAll()")
	fmt.Println(fh)

	data, err := ioutil.ReadAll(fh.f)

	return data, err
}

func (fh *FileHandler) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	fmt.Println("fh.Flush()")
	fmt.Println(fh)

	return nil
}

// func (fh *FileHandler) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
// 	fmt.Println("fh.Write()")
// 	data := req.Data
// 	offset := req.Offset
// 	n, err := fh.f.WriteAt(data, offset)
// 	if err != nil {
// 		logger.Warnf("Fail to write file for %v", err)
// 	}
// 	resp.Size = n
// 	return err
// }
 
// func (fh *FileHandler) Flush(ctx context.Context, req *fuse.FlushRequest) error {
// 	fmt.Println("fh.Flush()")
// 	err := fh.f.Sync()
// 	if err != nil {
// 		logger.Warnf("Fail to sync for %v", err)
// 	}

// 	return err
// }






