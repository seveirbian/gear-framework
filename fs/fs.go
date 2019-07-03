package fs

import (
	"os"
	"io"
	"fmt"
	"archive/tar"
	// "time"
	// "errors"
	// "reflect"
	"strings"
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
	monitorFlag = false
	logger = logrus.WithField("fs", "gearFS")

	GearPath             = "/var/lib/gear/"
	GearPublicCachePath  = filepath.Join(GearPath, "public")
)

type GearFS struct {
	MountPoint string

	IndexImagePath string
	PrivateCachePath string
	UpperPath string

	ManagerIp string
	ManagerPort string
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
	upperPath, err := ValidatePath(g.UpperPath)
	if err != nil {
		logrus.Fatalf("upperPath: %s is not valid...", g.UpperPath)
	}
	mountPoint, err := ValidatePath(g.MountPoint)
	if err != nil {
		logrus.Fatalf("mountPoint: %s is not valid...", g.MountPoint)
	}

	// 2. 在挂载点创建fuse连接
	c, err := fuse.Mount(mountPoint, fuse.AllowOther())
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
		err := c.Close()
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}(c)

	// 4. 初始化fuse文件系统
	filesys := Init(indexImagePath, privateCachePath, upperPath, g.ManagerIp, g.ManagerPort)

	// 5. 使用fuse文件系统服务挂载点的fuse连接
	if err := fuseFS.Serve(c, filesys); err != nil {
		fmt.Println(err)
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		fmt.Println(err)
	}
}

func (g * GearFS) StartAndNotify(notify chan int, monitor bool) {
	// 0. 设置检测位
	monitorFlag = monitor

	// 1. 检测index image目录、 private cache目录和挂载点目录是否合法
	indexImagePath, err := ValidatePath(g.IndexImagePath)
	if err != nil {
		logrus.Fatalf("indexImagePath: %s is not valid...", g.IndexImagePath)
	}
	privateCachePath, err := ValidatePath(g.PrivateCachePath)
	if err != nil {
		logrus.Fatalf("privateCachePath: %s is not valid...", g.PrivateCachePath)
	}
	upperPath, err := ValidatePath(g.UpperPath)
	if err != nil {
		logrus.Fatalf("upperPath: %s is not valid...", g.UpperPath)
	}
	mountPoint, err := ValidatePath(g.MountPoint)
	if err != nil {
		logrus.Fatalf("mountPoint: %s is not valid...", g.MountPoint)
	}

	// 2. 在挂载点创建fuse连接
	c, err := fuse.Mount(mountPoint, fuse.AllowOther())
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
		err := c.Close()
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}(c)

	// 4. 初始化fuse文件系统
	filesys := Init(indexImagePath, privateCachePath, upperPath, g.ManagerIp, g.ManagerPort)

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

func Init(indexImagePath, privateCachePath, upperPath, managerIp, managerPort string) *FS {
	// 对第二次启动实现加速
	// 1. 首先判断在gear-diff目录下是否存在RecordFiles文件
	_, err := os.Lstat(filepath.Join(indexImagePath, "RecordFiles"))
	if err != nil {
		// 不存在
	}else {
		// 存在
		_, err := os.Lstat(filepath.Join(indexImagePath, "prefetched"))
		if err != nil {
			b, err := ioutil.ReadFile(filepath.Join(indexImagePath, "RecordFiles"))
			if err != nil {
				logger.Warnf("Fail to read file for %v", err)
			}
			values := string(b)

			files := strings.Split(values, " ")
			files = files[0:len(files)-1]

			needToDownloadFiles := []string{}

			for _, file := range files {
				_, err := os.Lstat(filepath.Join(GearPublicCachePath, file))
				if err != nil {
					needToDownloadFiles = append(needToDownloadFiles, file)
				}
			}

			fmt.Println(needToDownloadFiles)

			v := url.Values{"files": needToDownloadFiles}


			resp, err := http.PostForm("http://"+managerIp+":"+managerPort+"/prefetch", v)
			if err != nil {
				logger.Warnf("Fail to prefetch for %v", err)
			}
			defer resp.Body.Close()

			f, err := os.Create(filepath.Join(GearPublicCachePath, "tmp"))
			if err != nil {
				logger.Fatalf("Fail to create tmp file for %V", err)
			}

			_, err = io.Copy(f, resp.Body)
			if err != nil {
				logger.Fatalf("Fail to copy for %v", err)
			}

			f.Close()

			f, err = os.Open(filepath.Join(GearPublicCachePath, "tmp"))
			if err != nil {
				logger.Warnf("Fail to open file for %v", err)
			}

			tr := tar.NewReader(f)

			for {
				th, err := tr.Next()
				if err == io.EOF {
					break;
				}

				tgt, err := os.Create(filepath.Join(GearPublicCachePath, th.Name))
				if err != nil {
					logger.Warnf("Fail to create tgt file for %v", err)
				}
				defer tgt.Close()

				_, err = io.Copy(tgt, tr)
				if err != nil {
					logger.Warnf("Fail to copy for %v", err)
				}

				err = os.Link(filepath.Join(GearPublicCachePath, th.Name), filepath.Join(privateCachePath, th.Name))
				if err != nil {
					logger.Fatalf("Fail to create hard link for %v", err)
				}

				// 设置文件权限
				err = os.Chmod(filepath.Join(privateCachePath, th.Name), 0777)
				if err != nil {
					logger.Warnf("Fail to chmod file for %v", err)
				}
			}

			f.Close()

			err = os.Remove(filepath.Join(GearPublicCachePath, "tmp"))
			if err != nil {
				logger.Warnf("Fail to remove tmp file for %v", err)
			}

			_, err = os.Create(filepath.Join(indexImagePath, "prefetched"))
			if err != nil {
				logger.Warnf("Fail to create file for %v", err)
			}
		}
	}

	return &FS{
		IndexImagePath: indexImagePath,
		PrivateCachePath: privateCachePath, 
		UpperPath: upperPath, 
	}
}

type FS struct {
	IndexImagePath string
	PrivateCachePath string
	UpperPath string
}

func (f *FS) Root() (fs.Node, error) {
	n := &Dir {
		// isRoot: true, 
		indexImagePath: f.IndexImagePath, 
		privateCachePath: f.PrivateCachePath, 
		upperPath: f.UpperPath, 

		relativePath: "/", 
	}

	return n, nil
}

type Dir struct {
	indexImagePath string
	privateCachePath string
	upperPath string

	relativePath string
}

// TODO: 实际获取每个目录的属性
func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	dirInfo, err := os.Lstat(filepath.Join(d.indexImagePath, d.relativePath))
	if err != nil {
		logger.Warnf("Fail to lstat dir.attr for %v", err)
	}

	attr.Inode = dirInfo.Sys().(*syscall.Stat_t).Ino
	attr.Size = uint64(dirInfo.Size())
	attr.Blocks = uint64(dirInfo.Sys().(*syscall.Stat_t).Blocks)
	attr.Mtime = dirInfo.ModTime()
	attr.Mode = dirInfo.Mode()
	attr.Nlink = uint32(dirInfo.Sys().(*syscall.Stat_t).Nlink)
	attr.Uid = dirInfo.Sys().(*syscall.Stat_t).Uid
	attr.Gid = dirInfo.Sys().(*syscall.Stat_t).Gid
	attr.BlockSize = uint32(dirInfo.Sys().(*syscall.Stat_t).Blksize)

	fmt.Println("\nd.Getattr")
	fmt.Println("d< ", d.relativePath, " >")
	fmt.Println("d.Attr< ", attr, " >")

	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var res []fuse.Dirent
	var path string

	path = filepath.Join(d.indexImagePath, d.relativePath)

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

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	target := filepath.Join(d.indexImagePath, d.relativePath, req.Name)

	fInfo, err := os.Lstat(target)
	if err != nil {
		logger.Warnf("Fail to Lstat target %s : %v", target, err)
		return nil, fuse.ENOENT
	}

	if fInfo.IsDir() {
		child := &Dir { 
			indexImagePath: d.indexImagePath, 
			privateCachePath: d.privateCachePath, 
			upperPath: d.upperPath, 
			relativePath: filepath.Join(d.relativePath, req.Name), 
		}
		return child, nil
	} else {
		if fInfo.Mode().IsRegular() {
			child := &File {
				isRegular: true, 
				indexImagePath: d.indexImagePath, 
				privateCachePath: d.privateCachePath, 
				upperPath: d.upperPath, 
				relativePath: filepath.Join(d.relativePath, req.Name), 
			}
			return child, nil
		} else {
			child := &File {
				isRegular: false, 
				indexImagePath: d.indexImagePath, 
				privateCachePath: d.privateCachePath, 
				upperPath: d.upperPath, 
				relativePath: filepath.Join(d.relativePath, req.Name), 
			}
			return child, nil
		}
	}

	return nil, fuse.ENOENT
}

func (d *Dir) Access(ctx context.Context, req *fuse.AccessRequest) error {
	return nil
}

type File struct {
	isRegular bool

	indexImagePath string
	privateCachePath string
	upperPath string

	relativePath string

	privateCacheName string
}

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	// 首先查看上层目录是否已经存在该文件
	upperFileInfo, err := os.Lstat(filepath.Join(f.upperPath, f.relativePath))
	if err == nil {
		// 是的话就返回upper目录的文件信息
		attr.Inode = upperFileInfo.Sys().(*syscall.Stat_t).Ino
		attr.Size = uint64(upperFileInfo.Size())
		attr.Blocks = uint64(upperFileInfo.Sys().(*syscall.Stat_t).Blocks)
		attr.Mtime = upperFileInfo.ModTime()
		attr.Mode = upperFileInfo.Mode()
		attr.Nlink = uint32(upperFileInfo.Sys().(*syscall.Stat_t).Nlink)
		attr.Uid = upperFileInfo.Sys().(*syscall.Stat_t).Uid
		attr.Gid = upperFileInfo.Sys().(*syscall.Stat_t).Gid
		attr.BlockSize = uint32(upperFileInfo.Sys().(*syscall.Stat_t).Blksize)
		fmt.Println("\nf.Getattr")
		fmt.Println("f< ", f.relativePath, " >")
		fmt.Print("f.attr< ", attr, ">\n")
		return nil
	}

	// 否则，再判断是否是普通文件，是否需要下载等等
	if f.isRegular {
		// 获取文件的cid
		name, err := ioutil.ReadFile(filepath.Join(f.indexImagePath, f.relativePath))
		if err != nil {
			logger.Warnf("Fail to read filename")
		}
		f.privateCacheName = string(name)

		// 检测private cache中是否存在该文件
		_, err = os.Lstat(filepath.Join(f.privateCachePath, f.privateCacheName))
		if err != nil {
			_, err := http.PostForm("http://localhost:2020"+"/get/"+f.privateCacheName, url.Values{"PATH":{f.privateCachePath}, "PERM":{"0777"}})
			if err != nil {
				logger.Warnf("Fail to get file for %v", err)
			}
			// 修改文件的权限
			IndexFileInfo, err := os.Lstat(filepath.Join(f.indexImagePath, f.relativePath))
			if err != nil {
				logger.Warnf("Fail to get index file info for %v", err)
			}
			err = os.Chmod(filepath.Join(f.privateCachePath, f.privateCacheName), IndexFileInfo.Mode())
			if err != nil {
				logger.Warnf("Fail to chmod for %v", err)
			}
			err = os.Chown(filepath.Join(f.privateCachePath, f.privateCacheName), int(IndexFileInfo.Sys().(*syscall.Stat_t).Uid), int(IndexFileInfo.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				logger.Warnf("Fail to chown for %v", err)
			}
		}

		fInfo, err := os.Lstat(filepath.Join(f.privateCachePath, f.privateCacheName))
		if err != nil {
			logger.Warnf("Fail to lstat file for %v", err)
		}

		attr.Size = uint64(fInfo.Size())

		IndexFileInfo, err := os.Lstat(filepath.Join(f.indexImagePath, f.relativePath))
		if err != nil {
			logger.Warnf("Fail to get index file info for %v", err)
		}

		attr.Inode = IndexFileInfo.Sys().(*syscall.Stat_t).Ino
		attr.Blocks = uint64(IndexFileInfo.Sys().(*syscall.Stat_t).Blocks)
		attr.Mtime = IndexFileInfo.ModTime()
		attr.Mode = IndexFileInfo.Mode()
		attr.Nlink = uint32(IndexFileInfo.Sys().(*syscall.Stat_t).Nlink)
		attr.Uid = IndexFileInfo.Sys().(*syscall.Stat_t).Uid
		attr.Gid = IndexFileInfo.Sys().(*syscall.Stat_t).Gid
		attr.BlockSize = uint32(IndexFileInfo.Sys().(*syscall.Stat_t).Blksize)
	} else {
		IndexFileInfo, err := os.Lstat(filepath.Join(f.indexImagePath, f.relativePath))
		if err != nil {
			logger.Warnf("Cannot stat file: %v", err)
		}

		attr.Inode = IndexFileInfo.Sys().(*syscall.Stat_t).Ino
		attr.Size = uint64(IndexFileInfo.Size())
		attr.Blocks = uint64(IndexFileInfo.Sys().(*syscall.Stat_t).Blocks)
		attr.Mtime = IndexFileInfo.ModTime()
		attr.Mode = IndexFileInfo.Mode()
		attr.Nlink = uint32(IndexFileInfo.Sys().(*syscall.Stat_t).Nlink)
		attr.Uid = IndexFileInfo.Sys().(*syscall.Stat_t).Uid
		attr.Gid = IndexFileInfo.Sys().(*syscall.Stat_t).Gid
		attr.BlockSize = uint32(IndexFileInfo.Sys().(*syscall.Stat_t).Blksize)
	}

	fmt.Println("\nf.Attr")
	fmt.Println("f< ", f.relativePath, " >")
	fmt.Println("f.Attr< ", attr, " >")

	go func() {
		if monitorFlag {
			image, err := os.Readlink(filepath.Join(f.indexImagePath, "gear-image"))
			if err != nil {
				logger.Warnf("Fail to readlink for %v", err)
			}
			resp, err := http.PostForm("http://localhost:2020/record/"+image+"/"+f.privateCacheName, url.Values{})
			if err != nil {
				logger.Warnf("Fail to post for %v", err)
			}
			defer resp.Body.Close()
		}
	}()

	return nil
}

func (f *File) Access(ctx context.Context, req *fuse.AccessRequest) error {
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {

	var fileHandler = FileHandler{}

	// 首先查看上层目录是否已经存在该文件
	_, err := os.Lstat(filepath.Join(f.upperPath, f.relativePath))
	if err == nil {
		// 是的话就打开upper目录的文件
		file, err := os.Open(filepath.Join(f.upperPath, f.relativePath))
		if err != nil {
			logger.Warnf("Fail to open file: %v", err)
		}
		fileHandler.f = file

		return &fileHandler, nil
	}

	// 否则，再判断是否是普通文件，是否需要下载等等
	if f.isRegular {
		// 1. 检查该镜像的私有缓存中是否存在cid文件
		_, err := os.Lstat(filepath.Join(f.privateCachePath, f.privateCacheName))
		if err != nil {
			// 该当前私有缓存中不存在cid文件，向gear client请求将cid文件下载到指定目录
			_, err := http.PostForm("http://localhost:2020"+"/get/"+f.privateCacheName, url.Values{"PATH":{f.privateCachePath}, "PERM":{"0777"}})
			if err != nil {
				logger.Warnf("Fail to get file for %v", err)
			}
			// 修改文件的权限
			IndexFileInfo, err := os.Lstat(filepath.Join(f.indexImagePath, f.relativePath))
			if err != nil {
				logger.Warnf("Fail to get index file info for %v", err)
			}
			err = os.Chmod(filepath.Join(f.privateCachePath, f.privateCacheName), IndexFileInfo.Mode())
			if err != nil {
				logger.Warnf("Fail to chmod for %v", err)
			}
			err = os.Chown(filepath.Join(f.privateCachePath, f.privateCacheName), int(IndexFileInfo.Sys().(*syscall.Stat_t).Uid), int(IndexFileInfo.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				logger.Warnf("Fail to chown for %v", err)
			}
		}

		// 2. 打开私有缓存中的文件
		file, err := os.Open(filepath.Join(f.privateCachePath, f.privateCacheName))
		if err != nil {
			logger.Warnf("Fail to open file: %v", err)
		}
		fileHandler.f = file

		return &fileHandler, nil
	}

	file, err := os.Open(filepath.Join(f.indexImagePath, f.relativePath))
	if err != nil {
		logger.Warnf("Fail to open file: %v", err)
	}
	fileHandler.f = file

	go func() {
		if monitorFlag {
			image, err := os.Readlink(filepath.Join(f.indexImagePath, "gear-image"))
			if err != nil {
				logger.Warnf("Fail to readlink for %v", err)
			}
			resp, err := http.PostForm("http://localhost:2020/record/"+image+"/"+f.privateCacheName, url.Values{})
			if err != nil {
				logger.Warnf("Fail to post for %v", err)
			}
			defer resp.Body.Close()
		}
	}()
	
	return &fileHandler, nil
}

func (f *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	target, err := os.Readlink(filepath.Join(f.indexImagePath, f.relativePath))
	if err != nil {
		logger.Warnf("Fail to read link for %v", err)
	}

	return target, err
}

type FileHandler struct {
	f *os.File
}

func (fh *FileHandler) Release(ctx context.Context, req *fuse.ReleaseRequest) error {

	return fh.f.Close()
}

func (fh *FileHandler) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {

	buf := make([]byte, req.Size)
	n, err := fh.f.Read(buf)
	resp.Data = buf[:n]

	return err
}

func (fh *FileHandler) ReadAll(ctx context.Context) ([]byte, error) {

	data, err := ioutil.ReadAll(fh.f)

	return data, err
}

func (fh *FileHandler) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return nil
}






