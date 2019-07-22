package graphdriver

import (
	// "github.com/docker/docker/pkg/archive"
	// "github.com/docker/docker/pkg/containerfs"
	// "log"
	// "context"
	"fmt"
	"io"
	"os"
	goPath "path"
	"time"
	"os/exec"
	"path/filepath"
	"sync"
	"net/http"
	"net/url"
	"strings"
	// "time"
	"archive/tar"
	"compress/gzip"
	"io/ioutil"
	// "strconv"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/seveirbian/gear/pkg"
	"github.com/seveirbian/gear/fs"
	"github.com/seveirbian/gear/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/containerfs"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/locker"
	"github.com/docker/docker/pkg/mount"
	"golang.org/x/sys/unix"
	// rsystem "github.com/opencontainers/runc/libcontainer/system"
	"github.com/docker/docker/daemon/graphdriver/overlay2"
	"github.com/docker/docker/daemon/graphdriver"
	// "github.com/docker/docker/pkg/directory"
	graphPlugin "github.com/docker/go-plugins-helpers/graphdriver"
	"github.com/opencontainers/selinux/go-selinux/label"
	"github.com/sirupsen/logrus"
)

var (
	GearPath             = "/var/lib/gear/"
	GearPrivateCachePath = filepath.Join(GearPath, "private")
	GearPublicCachePath  = filepath.Join(GearPath, "public")
	GearNFSPath          = filepath.Join(GearPath, "nfs")
	GearBuildPath        = filepath.Join(GearPath, "build")
	GearImagesPath       = filepath.Join(GearPath, "images")
	GearContainersPath   = filepath.Join(GearPath, "containers")
	GearPushPath         = filepath.Join(GearPath, "push")
)

var (
	// 监控gearfs的时间
	monitorTime = 600

	// 监测是否是第二次挂载容器层
	secondGet = map[string]bool{}

	// 监控gearfs关闭
	umountGearFs = map[string]chan time.Time{}
)

var (
	// untar defines the untar method
	untar = chrootarchive.UntarUncompressed
	// ApplyUncompressedLayer defines the unpack method used by the graph
	// driver
	ApplyUncompressedLayer = chrootarchive.ApplyUncompressedLayer

	gearCtr = map[string]int{}
	gearCommit = map[string]int{}
)

const (
	driverName = "geargraphdriver"
	linkDir    = "l"
	lowerFile  = "lower"
	maxDepth   = 128

	// idLength represents the number of random characters
	// which can be used to create the unique link identifier
	// for every layer. If this value is too long then the
	// page size limit for the mount command may be exceeded.
	// The idLength should be selected such that following equation
	// is true (512 is a buffer for label metadata).
	// ((idLength + len(linkDir) + 1) * maxDepth) <= (pageSize - 512)
	idLength = 26
)

type Driver struct {
	home             string
	uidMaps          []idtools.IDMap
	gidMaps          []idtools.IDMap
	ctr              *graphdriver.RefCounter
	naiveDiff        graphdriver.DiffDriver
	supportsDType    bool
	locker           *locker.Locker
	dockerDriver     graphdriver.Driver

	ManagerIp        string
	ManagerPort      string
	MonitorIp        string
	MonitorPort      string
}

var (
	logger                = logrus.WithField("storage-driver", "geargraphdriver")
	backingFs             = "<unknown>"
	projectQuotaSupported = false

	useNaiveDiffLock sync.Once
	useNaiveDiffOnly bool

	indexOff string
)

// 初始化驱动
func (d *Driver) Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error {
	fmt.Println("\nInit func parameters: ")
	fmt.Printf("  home: %s\n", home)
	fmt.Printf("  options: %s\n", options)
	fmt.Println("  uidMaps: ", uidMaps)
	fmt.Println("  gidMaps: ", gidMaps)

	d.uidMaps = uidMaps
	d.gidMaps = gidMaps
	d.home = home

	driver, err := overlay2.Init(home, []string{}, nil, nil)
	if err != nil {
		logger.WithField("err", err).Warn("Fail to create overlay2 driver...")
		return err
	}
	d.dockerDriver = driver
	// d.naiveDiff = graphdriver.NewNaiveDiffDriver(d, uidMaps, gidMaps)

	return nil
}

// 为镜像层创建目录
func (d *Driver) Create(id, parent, mountlabel string, storageOpt map[string]string) (retErr error) {
	fmt.Println("\nCreate func parameters: ")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)
	fmt.Printf("  mountLabel: %s\n", mountlabel)
	fmt.Printf("  storageOpt: %s\n", storageOpt)

	// 使用d.dockerDriver为docker镜像和gear镜像创建镜像层文件夹
	retErr = d.dockerDriver.Create(id, parent, &graphdriver.CreateOpts{
		MountLabel: mountlabel, 
		StorageOpt:storageOpt, 
	})

	return 
}

// 为容器层创建目录
func (d *Driver) CreateReadWrite(id, parent, mountlabel string, storageOpt map[string]string) (retErr error) {
	fmt.Println("\nCreateReadWrite func parameters: ")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)
	fmt.Printf("  mountLabel: %s\n", mountlabel)
	fmt.Printf("  storageOpt: %s\n", storageOpt)

	retErr = d.dockerDriver.CreateReadWrite(id, parent, &graphdriver.CreateOpts{
		MountLabel: mountlabel, 
		StorageOpt:storageOpt, 
	})

	// 1. 检测parent的目录中是否有gear-lower
	path := filepath.Join(d.home, parent, "gear-lower")
	_, err := os.Lstat(path)
	if err == nil {
		// 2. 检测gear镜像目录中是否有gear-lower
		gearLower := filepath.Join(d.home, parent, "gear-lower")
		target, err := os.Readlink(gearLower)
		if err != nil {
			logger.Fatal("No gear-lower...")
		}

		// 3. 在孩子目录中也复制一份gear-lower
		childGearLink := filepath.Join(d.home, id, "gear-lower")
		err = os.Symlink(target, childGearLink)
		if err != nil {
			logger.Warnf("Fail to read gear-lower for %v", err)
		}
	}

	return
}

// 删除id目录
func (d *Driver) Remove(id string) error {
	fmt.Printf("\nRemove func parameters: \n")
	fmt.Printf("  id: %s\n", id)

	err := d.dockerDriver.Remove(id)
	return err
}

// Get creates and mounts the required file system for the given id and returns the mount path
func (d *Driver) Get(id, mountLabel string) (containerfs.ContainerFS, error) {
	fmt.Printf("\nGet func parameters: \n")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  mountlabel: %s\n", mountLabel)

	// 1. 检测目录中是否有gear-lower
	path := filepath.Join(d.home, id, "gear-lower")
	gearPath, err := os.Readlink(path)
	if err != nil {
		containerFs, err := d.dockerDriver.Get(id, mountLabel)

		return containerFs, err
	} else {
		// 检测该目录下是否有lower文件
		_, err := os.Lstat(filepath.Join(d.home, id, "lower"))
		if err != nil {
			// 有gear-lower但没有lower的是gear镜像
			// 当前目录是gear镜像目录，不需要gear fs的帮助
			// 直接将gear-diff目录返回？？？
			gearDiffDir := filepath.Join(d.home, id, "gear-diff")
			return containerfs.NewLocalContainerFS(gearDiffDir), nil
		} else {
			// 既有gear-lower也有lower的是gear容器目录
			// 当前目录是gear容器目录，需要gear fs的挂载
			// 2. 有，需要gear fs目录的挂载
			gearDiffDir := filepath.Join(gearPath, "diff")
			gearGearDir := filepath.Join(gearPath, "gear-diff")

			// 3. 查找gear-diff目录下，gear-image软链接的镜像名和tag
			gearImage, err := os.Readlink(filepath.Join(gearGearDir, "gear-image"))
			if err != nil {
				logger.Warnf("Fail to read gear-image symlink for %v", err)
			}

			// 4. 获取镜像自己的私有cache
			gearImagePrivateCache := filepath.Join(GearPrivateCachePath, gearImage)
			_, err = os.Lstat(gearImagePrivateCache)
			if err != nil {
				// 创建一个
				err := os.MkdirAll(gearImagePrivateCache, 0700)
				if err != nil {
					logger.Warnf("Fail to create image private cache dir for %v", err)
				}
			}
			
			// 判断是否需要监控gearfs
			needMonitor := false

			recordChan := make(chan types.MonitorFile, 100)
			// recordFilesChan := make(chan string, 100)
			// recordFileNamesChan := make(chan string, 100)

			initLayerPath := ""
			if !strings.Contains(id, "-init") {
				// 读取镜像层中lower文件内容
				data, err := ioutil.ReadFile(filepath.Join(d.home, id, "lower"))
				if err != nil {
					logger.Warnf("Fail to read lower file for %v", err)
				}
				// 获取-init目录
				dataSlices := strings.Split(string(data), ":")
				lPath := dataSlices[0]
				initLayerDir, err := os.Readlink(filepath.Join(d.home, lPath))
				initLayerPath = filepath.Join(d.home, id, initLayerDir)
				fmt.Println(initLayerPath)
			}

			_, err = os.Lstat(filepath.Join(gearPath, "gear-diff", "RecordFiles"))
			if err != nil {
				// 需要监控该镜像
				logger.Warnf("Monitoring...")
				needMonitor = true

				recordFileNames := []string{}
				recordFiles := []string{}

				go func(id string) {
					// 判断是否需要监测
					if strings.Contains(id, "-init") {
						return 
					}
					if _, ok := secondGet[id]; !ok {
						secondGet[id] = true
						return
					}

					fmt.Println("Monitoring...")
					t := time.NewTimer(time.Duration(monitorTime) * time.Second)

					reportChan := make(chan time.Time)
					if _, ok := umountGearFs[id]; !ok {
						umountGearFs[id] = reportChan
					}

					dupFiles := map[string]bool{}

					for {
						select {
						case file := <- recordChan:
							if _, ok := dupFiles[file.Hash]; !ok {
								dupFiles[file.Hash] = true
								recordFiles = append(recordFiles, file.Hash)
								recordFileNames = append(recordFileNames, file.RelativePath)
							}
							// TODO: 对于软连接或者多个文件指向同一个哈希文件的情况
						case <- t.C:
							// 向manager汇报
							v := url.Values{"files": recordFiles, "filenames": recordFileNames, "image": []string{gearImage}}

							// 在本地创建RecordFiles文件
							if len(recordFiles) != len(recordFileNames) {
								logger.Warnf("Something went error that len(recordFiles) != len(recordFileNames)")
							} else {
								content := ""
								for i := 0; i < len(recordFiles) - 1; i++ {
									content = content+recordFileNames[i]+" "+recordFiles[i]+"\n"
								}
								content = content+recordFileNames[len(recordFiles)-1]+" "+recordFiles[len(recordFiles)-1]

								err = ioutil.WriteFile(filepath.Join(gearPath, "gear-diff", "RecordFiles"), []byte(content), os.ModePerm)
								if err != nil {
									logger.Warnf("Fail to write recordfiles for %v", err)
								}

							}

							resp, err := http.PostForm("http://"+d.MonitorIp+":"+d.MonitorPort+"/event", v)
							if err != nil {
								logger.Warnf("Fail to report to monitor for %v", err)
							}

							fmt.Println("send event!")

							defer resp.Body.Close()

							delete(umountGearFs, id)

							return
						case <- umountGearFs[id]:
							// 向manager汇报
							v := url.Values{"files": recordFiles, "filenames": recordFileNames, "image": []string{gearImage}}

							// 在本地创建RecordFiles文件
							if len(recordFiles) != len(recordFileNames) {
								logger.Warnf("Something went error that len(recordFiles) != len(recordFileNames)")
							} else {
								content := ""
								for i := 0; i < len(recordFiles) - 1; i++ {
									content = content+recordFileNames[i]+" "+recordFiles[i]+"\n"
								}
								content = content+recordFileNames[len(recordFiles)-1]+" "+recordFiles[len(recordFiles)-1]

								err = ioutil.WriteFile(filepath.Join(gearPath, "gear-diff", "RecordFiles"), []byte(content), os.ModePerm)
								if err != nil {
									logger.Warnf("Fail to write recordfiles for %v", err)
								}
							}

							resp, err := http.PostForm("http://"+d.MonitorIp+":"+d.MonitorPort+"/event", v)
							if err != nil {
								logger.Warnf("Fail to report to monitor for %v", err)
							}

							fmt.Println("send event!")

							defer resp.Body.Close()

							delete(umountGearFs, id)

							return
						}
					}
				}(id)
			} else {
				if !strings.Contains(id, "-init") {
					// 判断是否存在prefetched文件
					files := []string{}
					tamplate := map[string]string{}

					b, err := ioutil.ReadFile(filepath.Join(gearPath, "gear-diff", "RecordFiles"))
					if err != nil {
						logger.Warnf("Fail to read file for %v", err)
					}
					values := string(b)	
					nameAndFiles := strings.Split(values, "\n")

					tmp := []string{}

					for _, nameAndFile := range nameAndFiles {
						c := strings.Split(nameAndFile, " ")
						if c[1] != "" {
							tmp = append(tmp, c[1])

							if _, ok := tamplate[c[1]]; !ok {
								tamplate[c[1]] = c[0]
							}
						}						
					}
					files = tmp

					_, err = os.Lstat(filepath.Join(gearPath, "gear-diff", "prefetched"))
					if err != nil {
						t := time.Now()

						needToDownloadFiles := []string{}

						for _, file := range files {
							_, err := os.Lstat(filepath.Join(GearPublicCachePath, file))
							if err != nil {
								needToDownloadFiles = append(needToDownloadFiles, file)
							}
						}

						v := url.Values{"files": needToDownloadFiles, "image": []string{gearImage}}

						resp, err := http.PostForm("http://"+d.ManagerIp+":"+d.ManagerPort+"/prefetch", v)
						if err != nil {
							logger.Warnf("Fail to prefetch for %v", err)
						}
						defer resp.Body.Close()

						tr := tar.NewReader(resp.Body)

						for {
							th, err := tr.Next()
							if err == io.EOF {
								break;
							}

							tgt, err := os.Create(filepath.Join(GearPublicCachePath, th.Name))
							if err != nil {
								logger.Warnf("Fail to create tgt file for %v", err)
							}

							gr, err := gzip.NewReader(tr)

							_, err = io.Copy(tgt, gr)
							if err != nil {
								logger.Fatalf("Fail to copy for %v", err)
							}

							gr.Close()
							tgt.Close()

							err = os.Link(filepath.Join(GearPublicCachePath, th.Name), filepath.Join(gearImagePrivateCache, th.Name))
							if err != nil {
								logger.Fatalf("Fail to create hard link for %v", err)
							}

							// 设置文件权限
							err = os.Chmod(filepath.Join(gearImagePrivateCache, th.Name), 0777)
							if err != nil {
								logger.Warnf("Fail to chmod file for %v", err)
							}
						}

						_, err = os.Create(filepath.Join(gearGearDir, "prefetched"))
						if err != nil {
							logger.Warnf("Fail to create file for %v", err)
						}

						fmt.Println("Time used: ", time.Since(t))
					} 

					// 将文件link到-init层目录
					lt := time.Now()
					fmt.Println(lt)
					if initLayerPath != "" {
						// 将文件link到-init层目录
						for _, file := range files {
							relativePath, ok := tamplate[file]
							if !ok {
								continue
							}
							_, err = os.Lstat(filepath.Join(initLayerPath, relativePath))
							if err != nil {
								initDir := goPath.Dir(filepath.Join(initLayerPath, relativePath))
								_, err = os.Lstat(initDir)
								if err != nil {
									err := os.MkdirAll(initDir, os.ModePerm)
									if err != nil {
										logger.Warnf("Fail to create initDir for %v", err)
									}
								}
								err = os.Link(filepath.Join("/var/lib/gear/public", file), filepath.Join(initLayerPath, relativePath))
								if err != nil {
									logger.Fatalf("Fail to create hard link for %v", err)
								}
								err = os.Chmod(filepath.Join(initLayerPath, relativePath), 0777)
								if err != nil {
									logger.Warnf("Fail to chmod file for %v", err)
								}
							}
						}
					}
					fmt.Println(time.Since(lt))
				}
			}

			// 5. 将/gear目录使用gear fs挂载到/diff目录下
			gearFS := &fs.GearFS {
				MountPoint: gearDiffDir, 
				IndexImagePath: gearGearDir, 
				PrivateCachePath: gearImagePrivateCache, 
				UpperPath: filepath.Join(d.home, id, "diff"), 

				ManagerIp: d.ManagerIp, 
				ManagerPort: d.ManagerPort, 

				RecordChan: recordChan, 

				InitLayerPath: initLayerPath, 
			}

			notify := make(chan int)


			// 判断是否是第一次挂载
			if _, ok := gearCtr[gearDiffDir]; !ok {
				// 第一次挂载
				gearCtr[gearDiffDir] = 1
				go gearFS.StartAndNotify(notify, needMonitor)
				<- notify
			} else {
				gearCtr[gearDiffDir] += 1
			}

			containerFs, err := d.dockerDriver.Get(id, mountLabel)

			return containerFs, err
		}
	}
}

// Put unmounts the mount path created for the give id.
// It also removes the 'merged' directory to force the kernel to unmount the
// overlay mount in other namespaces.
func (d *Driver) Put(id string) error {
	fmt.Printf("\nPut func parameters: \n")
	fmt.Printf("  id: %s\n", id)

	t := time.Now()
	Eterr := d.dockerDriver.Put(id)
	fmt.Println("Put time: ", time.Since(t))

	// 1. 检测孩子的diff目录中是否有gear-image软链接
	path := filepath.Join(d.home, id, "gear-lower")
	gearPath, err := os.Readlink(path)
	if err != nil {
		return Eterr
	} else {
		// 2. 有，卸载diff目录
		gearDiffDir := filepath.Join(gearPath, "diff")

		if ctr, _ := gearCtr[gearDiffDir]; ctr <= 1 {
			fmt.Println("该镜像只有一个容器！")
			delete(gearCtr, gearDiffDir)

			fmt.Println("删除umountgearfs中对应的id词条！")
			if _, ok := umountGearFs[id]; ok {
				umountGearFs[id] <- time.Now()
			}

			fmt.Println("卸载gearfs！")
			cmd := exec.Command("umount", gearDiffDir)
			err := cmd.Run()
			if err != nil {
				logger.Warnf("Fail to umount diff for %v", err)
			}
			// 3. 强制删除diff目录
			err = os.RemoveAll(gearDiffDir)
			if err != nil {
				logger.Warnf("Fail to remove diff dir for %v", err)
			}
			// 4. 新建diff目录
			err = os.MkdirAll(gearDiffDir, 0700)
			if err != nil {
				logger.Warnf("Fail to create diff dir for %v", err)
			}
			fmt.Println("Put完成！")
		} else {
			gearCtr[gearDiffDir] -= 1
		}

		return nil
	}
}

// 查看id是否已经被挂载了
func (d *Driver) Exists(id string) bool {
	fmt.Printf("\nExists func parameters: \n")
	fmt.Printf("  id: %s\n", id)

	_, err := os.Stat(goPath.Join(d.home, id))
	return err == nil
}

// Status returns current driver information in a two dimensional string array.
// Output contains "Backing Filesystem" used in this implementation
func (d *Driver) Status() [][2]string {
	fmt.Printf("\nStatus func parameters: \n")

	return [][2]string{
		{"Backing Filesystem", backingFs},
		// {"Supports d_type", strconv.FormatBool(d.supportsDType)},
		// {"Native Overlay Diff", strconv.FormatBool(!useNaiveDiff(d.home))},
	}
}

// GetMetadata returns metadata about the overlay driver such as the LowerDir,
// UpperDir, WorkDir, and MergeDir used to store data
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	fmt.Printf("\nGetMetadata func parameters: \n")
	fmt.Printf("  id: %s\n", id)

	metadata, err := d.dockerDriver.GetMetadata(id)

	return metadata, err
}

// Cleanup any state created by overlay which should be cleaned when daemon
// is being shutdown. For now, we just have to unmount the bind mounted
// we had created
func (d *Driver) Cleanup() error {
	fmt.Printf("\nCleanup func parameters: \n")

	return mount.Unmount(d.home)
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be ""
func (d *Driver) Diff(id, parent string) io.ReadCloser {
	fmt.Printf("\nDiff func parameters: \n")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)
	
	// 检测当前镜像是否是gear镜像
	path := filepath.Join(d.home, id, "gear-lower")
	_, err := os.Lstat(path)
	if err != nil {
		// 不是gear镜像
		diff, err := d.dockerDriver.Diff(id, parent)
		if err != nil {
			logger.Warnf("Fail to diff for %v",  err)
		}
		return diff
	} else {
		// 是gear镜像
		// 1. 将该层目录下的文件构建成gear镜像的目录树，即：将普通文件内容替换成内容的哈希值，内容以哈希值
		// 命名上传到remote storage
		currentDir := filepath.Join(d.home, id, "diff")
		pushDir := filepath.Join(GearPushPath, id)
		err := os.MkdirAll(pushDir, os.ModePerm)
		if err != nil {
			logger.Warnf("Fail to make push dir for %v", err)
		}

		err = filepath.Walk(currentDir, func(path string, f os.FileInfo, err error) error {
			if f == nil {
	    		return err
	    	}

	    	// 防止本目录也处理了
			pathSlice := strings.SplitAfter(path, currentDir)
			if pathSlice[1] == "" {
				return nil
			}

			if f.Mode().IsRegular() {
				hashValue := []byte(pkg.HashAFileInMD5(path))

				src, err := os.Open(path)
				if err != nil {
					logger.Warnf("Fail to open file: %s\n", path)
					return err
				}
				defer src.Close()

				dst, err := os.Create(filepath.Join(pushDir, string(hashValue)))
				if err != nil {
					logger.Warnf("Fail to create file: %s\n", filepath.Join(pushDir, string(hashValue)))
					return err
				}
				defer dst.Close()

				// 拷贝文件内容
				_, err = io.Copy(dst, src)
				if err != nil {
					logger.Warn("Fail to copy file...")
					return err
				}

				err = ioutil.WriteFile(path, hashValue, f.Mode().Perm())
				if err != nil {
					logger.Warnf("Fail to write file for %v", err)
				}
			}

			return nil
		})

		if err != nil {
			logger.Warnf("Fail to walk id dir for %v", err)
		}

		resp, err := http.PostForm("http://localhost:2020/upload", url.Values{"PATH": {pushDir}})
		if err != nil {
			logger.Warnf("Fail to post for %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warnf("Fail to push files...")
		}

		err = os.RemoveAll(pushDir)
		if err != nil {
			logger.Warnf("Fail to remove push dir for %v", err)
		}

		// 2. 联合挂载，对上下层目录树做Diff
		parent, err := os.Readlink(filepath.Join(d.home, id, "gear-lower"))
		if err != nil {
			logger.Warnf("Fail to readlink for %v", err)
		}
		parentDir := filepath.Join(parent, "gear-diff")
		opts := "lowerdir=" + parentDir + ",upperdir=" + currentDir + ",workdir=" + filepath.Join(d.home, id, "work")
		mountData := label.FormatMountLabel(opts, "")
		mount := unix.Mount
		mergedDir := filepath.Join(d.home, id, "merged")
		mountTarget := mergedDir
		rootUID, rootGID, err := idtools.GetRootUIDGID(d.uidMaps, d.gidMaps)
		if err != nil {
			logger.Fatalf("Fail to get rootUIDGID for %v", err)
		}
		if err := idtools.MkdirAndChown(mergedDir, 0700, idtools.Identity{UID: rootUID, GID: rootGID}); err != nil {
			logger.Fatalf("Fail to mk merged dir for %v", err)
		}
		if err := mount("overlay", mountTarget, "overlay", 0, mountData); err != nil {
			logger.Fatalf("Fail to mount overlay for %v", err)
		}
		if err := os.Chown(filepath.Join(filepath.Join(d.home, id, "work"), "work"), rootUID, rootGID); err != nil {
			logger.Fatalf("Fail to get chown for %v", err)
		}

		layerFs := mountTarget

		archive, err := archive.Tar(layerFs, archive.Uncompressed)
		if err != nil {
			logger.Fatalf("Fail to get tar for %v", err)
		}
		return ioutils.NewReadCloserWrapper(archive, func() error {
			err := archive.Close()
			if err := unix.Unmount(mountTarget, unix.MNT_DETACH); err != nil {
				logger.Debugf("Failed to unmount %s overlay: %s - %v", id, mountTarget, err)
			}
			if err := unix.Rmdir(mountTarget); err != nil && !os.IsNotExist(err) {
				logger.Debugf("Failed to remove %s overlay: %v", id, err)
			}
			return err
		})
	}
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes
func (d *Driver) Changes(id, parent string) ([]graphPlugin.Change, error) {
	fmt.Printf("\nChanges func parameters: \n")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)

	dockerChanges, err := d.dockerDriver.Changes(id, parent)
	if err != nil {
		return []graphPlugin.Change{}, err
	}

	var changes = []graphPlugin.Change{}
	for _, change := range dockerChanges {
		changes = append(changes, graphPlugin.Change{
			Path: change.Path, 
			Kind: graphPlugin.ChangeKind(change.Kind), 
		})
	}

	return changes, nil
}

// ApplyDiff applies the new layer into a root
// docker pull will call this func
// TODO: this func has bug
func (d *Driver) ApplyDiff(id, parent string, diff io.Reader) (int64, error) {
	fmt.Printf("\nApplyDiff func parameters: \n")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)

	// 1. 直接将数据解压到diff文件中
	size, err := d.dockerDriver.ApplyDiff(id, parent, diff)

	// 4. 检测当前镜像是否是gear镜像
	path := filepath.Join(d.home, id, "diff", "gear-image")
	_, err = os.Lstat(path)

	// 判断gear镜像
	if err != nil {
		// 不是gear镜像
		// 直接返回
		return size, nil
	} else {
		// 是gear镜像，将diff文件夹重命名为gear-diff
		err := os.Rename(filepath.Join(d.home, id, "diff"), filepath.Join(d.home, id, "gear-diff"))
		if err != nil {
			logger.Warnf("Fail to rename diff dir to gear-diff for %v", err)
		}

		// 2. 创建gearlink
		gearLower := filepath.Join(d.home, id, "gear-lower")
		err = os.Symlink(filepath.Join(d.home, id), gearLower)
		if err != nil {
			logger.Warnf("Fail to create gear link for %v", err)
		}

		// 3. 删除overlay driver创建的lower文件
		_, err = os.Lstat(filepath.Join(d.home, id, "lower"))
		if err == nil {
			err := os.Remove(filepath.Join(d.home, id, "lower"))
			if err != nil {
				logger.Warnf("Fail to remove lower for %v", err)
			}
		}

		// 4. 创建diff文件夹
		err = os.MkdirAll(filepath.Join(d.home, id, "diff"), os.ModePerm)
		if err != nil {
			logger.Warnf("Fail to mkdir diff for %v", err)
		}

		return size, nil
	}
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory
func (d *Driver) DiffSize(id, parent string) (int64, error) {
	fmt.Printf("\nDiffSizee func parameters: \n")
	fmt.Printf("  id: %s\n", id)
	fmt.Printf("  parent: %s\n", parent)

	return d.dockerDriver.DiffSize(id, parent)
}

func (d *Driver) Capabilities() graphdriver.Capabilities {
	fmt.Printf("\nCapabilities func parameters: \n")

	return graphdriver.Capabilities{
		ReproducesExactDiffs: false,
	}
}

// protoDriver需要实现该方法
func (d *Driver) String() string {
	return "Gear Graphdriver"
}
