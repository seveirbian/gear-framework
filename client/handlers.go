package client

import (
	"io"
	"os"
	"fmt"
	"sync"
	// "strings"
	// "syscall"
	"strconv"
	"net/url"
	"net/http"
	// "encoding/json"
	"path/filepath"
	"github.com/labstack/echo"
	"github.com/seveirbian/gear/push"
	// "github.com/seveirbian/gear/pkg"
	// "github.com/seveirbian/gear/types"
)

var (
	recordedImages = map[string]bool{}
	imageFetchedFiles = map[string]map[string]bool{}
)

func handleInfo(c echo.Context) error {
	resp := strconv.FormatUint(cli.Self.ID, 10)+":"+cli.Self.IP+":"+cli.Self.Port

	return c.String(http.StatusOK, resp)
}

func handleDownload(c echo.Context) error {
	// 1. 获取请求的OID
	cid := c.Param("CID")

	// 2. 查看本地缓存是否存在
	_, err := os.Stat(filepath.Join("/var/lib/gear/public", cid))

	var l = sync.RWMutex{}

	// 不存在
	if err != nil {
		// 从manager节点下载cid文件
		resp, err := http.PostForm("http://"+cli.Manager.IP+":"+cli.Manager.Port+"/pull/"+cid, url.Values{})
		if err != nil {
			logger.Warnf("Fail to pull from manager for %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return c.NoContent(resp.StatusCode)
		}

		l.Lock()

		dst, err := os.Create(filepath.Join("/var/lib/gear/public", cid))
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to create sharing file...")
		}
		defer dst.Close()

		_, err = io.Copy(dst, resp.Body)
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to copy file...")
		}

		l.Unlock()
	}

	l.RLock()
	defer l.RUnlock()

	return c.Attachment(filepath.Join("/var/lib/gear/public", cid), cid)
}

func handleGet(c echo.Context) error {
	cid := c.Param("CID")
	cidPath := c.FormValue("PATH")

	// 1. 搜索所有集群节点，找到最优节点获取文件
	// cidInUint64, _ := strconv.ParseUint(cid, 10, 64)
	// var distance uint64 = cidInUint64 ^ cli.Self.ID
	// candidata := cli.Self

	// cli.NodesMu.RLock()
	// for cliId, client := range(cli.Nodes) {
	// 	if distance > (cidInUint64 ^ cliId) {
	// 		distance = cidInUint64 ^ cliId
	// 		candidata = client
	// 	}
	// }
	// cli.NodesMu.RUnlock()

	// 2. 确认本地是否存在文件
	_, err := os.Lstat(filepath.Join("/var/lib/gear/public", cid))
	if err == nil {
		// 跳过下载步骤
		// 创建硬连接到镜像私有缓存目录下
		err = os.Link(filepath.Join("/var/lib/gear/public", cid), filepath.Join(cidPath, cid))
		if err != nil {
			logger.Fatalf("Fail to create hard link for %v", err)
		}

		// 设置文件权限
		err = os.Chmod(filepath.Join(cidPath, cid), 0777)
		if err != nil {
			logger.Warnf("Fail to chmod file for %v", err)
		}

		return c.NoContent(http.StatusOK)
	}

	// 从manager节点下载cid文件
	resp, err := http.PostForm("http://"+cli.Manager.IP+":"+cli.Manager.Port+"/pull/"+cid, url.Values{})
	if err != nil {
		logger.Warnf("Fail to pull from manager for %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.NoContent(resp.StatusCode)
	}

	f, err := os.Create(filepath.Join("/var/lib/gear/public", cid))
	if err != nil {
		logger.Fatalf("Fail to create file for %V", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		logger.Fatalf("Fail to copy for %v", err)
	}

	// 4. 创建硬连接到镜像私有缓存目录下
	err = os.Link(filepath.Join("/var/lib/gear/public", cid), filepath.Join(cidPath, cid))
	if err != nil {
		logger.Fatalf("Fail to create hard link for %v", err)
	}

	// 5. 设置文件权限
	err = os.Chmod(filepath.Join(cidPath, cid), 0777)
	if err != nil {
		logger.Warnf("Fail to chmod file for %v", err)
	}

	return c.NoContent(http.StatusOK)
}

func handleUpload(c echo.Context) error {
	// 1. get imageName and tag
	dir := c.FormValue("PATH")
	fmt.Println(dir)
	
	pusher, err := push.InitPusher(dir, cli.Manager.IP, cli.Manager.Port, false)
    if err != nil {
        logger.Fatal("Fail to init a pusher to push gear image...")
    }

    pusher.Push()

    return c.NoContent(http.StatusOK)
}

func handleRecorded(c echo.Context) error {
	image := c.Param("IMAGE")
	if _, ok := recordedImages[image]; !ok {
		fmt.Println("IMAGE: ", image)
		fmt.Println("Recorded: ", ok)

		recordedImages[image] = true
		imageFetchedFiles[image] = map[string]bool{}
		return c.NoContent(http.StatusOK)
	}

	return c.NoContent(http.StatusCreated)
}

func handleRecord(c echo.Context) error {
	// 1. get IMAGE and CID
	image := c.Param("IMAGE")
	cid := c.Param("CID")

	if notReported, ok := recordedImages[image]; ok && !notReported {
		return c.NoContent(http.StatusOK)
	}

	fmt.Println(cid)
	
	// 2. 检测是否已经记录过该镜像
	if _, ok := recordedImages[image]; ok {
		if _, ok := imageFetchedFiles[image][cid]; !ok {
			imageFetchedFiles[image][cid] = true
		}
	}

    return c.NoContent(http.StatusOK)
}

func handleReport(c echo.Context) error {
	image := c.Param("IMAGE")

	if notReported, ok := recordedImages[image]; ok && notReported {
		recordedImages[image] = false
	} else {
		return c.NoContent(http.StatusOK)
	}

	// 向manager汇报
	files := []string{}
	for file, _ := range imageFetchedFiles[image] {
		files = append(files, file)
	}
	v := url.Values{"files": files}

	fmt.Println(v)

	resp, err := http.PostForm("http://"+cli.Monitor.IP+":"+cli.Monitor.Port+"/event/"+image, v)
	if err != nil {
		logger.Warnf("Fail to report to monitor for %v", err)
	}
	defer resp.Body.Close()
	
	return c.NoContent(resp.StatusCode)
}









