package manager

import (
	"os"
	"io"
	"fmt"
	// "syscall"
	"path/filepath"
	"strconv"
	"net/http"
	"github.com/labstack/echo"
	"github.com/seveirbian/gear/pkg"
	"github.com/seveirbian/gear/types"
)

var (
	GearPath             = "/var/lib/gear/"
	GearStoragePath      = filepath.Join(GearPath, "storage")
)

func handleNodes(c echo.Context) error {
	var resp string

	mgr.NodesMu.RLock()
 
	for _, node := range(mgr.Nodes) {
		resp = resp + strconv.FormatUint(node.ID, 10) + ":" + node.IP + ":" + node.Port + ";"
	}

	defer mgr.NodesMu.RUnlock()
	
	return c.String(http.StatusOK, resp)
}

func handleJoin(c echo.Context) error {
	mIP := c.Param("IP")
	mPort := c.Param("Port")

	id := pkg.CreateIdFromIP(mIP)

	// 对Nodes数据结构加读锁
	mgr.NodesMu.RLock()

	_, ok := mgr.Nodes[id]

	// 解锁
	mgr.NodesMu.RUnlock()

	if !ok {
		// 对Nodes数据结构加写锁
		mgr.NodesMu.Lock()

		mgr.Nodes[id] = types.Node{
			ID:   id,
	        IP:   mIP,
	        Port: mPort,
		}

		// 解锁
		mgr.NodesMu.Unlock()
	}

	return c.NoContent(http.StatusOK)
}

func handlePull(c echo.Context) error {
	cid := c.Param("CID")

	fmt.Println(filepath.Join(GearStoragePath, cid))

	_,  err := os.Lstat(filepath.Join(GearStoragePath, cid))
	if err != nil {
		logger.Warnf("Fail to lstat file for %v", err)
	}

	// 返回本地文件
	// f, err := os.Open(filepath.Join(GearStoragePath, cid))
	// if err != nil {
	// 	logger.Fatalf("Fail to open file: %s\n", filepath.Join(GearStoragePath, cid))
	// }
	// defer f.Close()
	// 上共享文件锁
	// err = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
	// if err != nil {
	// 	logger.Fatal("Fail to lock file in a sharing way...")
	// }
	err = c.Attachment(filepath.Join(GearStoragePath, cid), cid)
	if err != nil {
		logger.Fatal("Fail to return file...")
	}
	// 解锁
	// err = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	// if err != nil {
	// 	logger.Fatal("Fail to unlock file in a sharing way...")
	// }

	return nil
}

func handleQuery(c echo.Context) error {
	cid := c.Param("CID")

	fmt.Printf("Querying %s\n", cid)

	_, err := os.Lstat(filepath.Join(GearStoragePath, cid))
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusOK)
}

func handlePush(c echo.Context) error {
	cid := c.Param("CID")
	file, err := c.FormFile("file")
	if err != nil {
		logger.Warnf("Fail to get formfile for %v", err)
	}

	// 检测是否已经存在cid文件
	_,  err = os.Lstat(filepath.Join(GearStoragePath, cid))
	if err != nil {
		// storage端没有cid文件，下载到storage
		src, err := file.Open()
		if err != nil {
			logger.Warnf("Fail to open file for %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(GearStoragePath, cid))
		if err != nil {
			logger.Warnf("Fail to create file for %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			logger.Warnf("Fail to copy for %V", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	// storage已经有cid文件，直接返回
	return c.NoContent(http.StatusOK)
}

func handleReport(c echo.Context) error {
	// 1. 获取镜像名
	image := c.Param("IMAGE")
	// 2. 获取文件
	values, err := c.FormParams()
	if err != nil {
		logger.Warnf("Fail to parse files for %v", err)
	}

	files := values["files"]

	fmt.Println(image)
	fmt.Println(files)

	return c.NoContent(http.StatusOK)
}






