package push

import (
	"fmt"
	"io"
	"os"
	"bytes"
	"strings"
	"net/url"
	"net/http"
	"mime/multipart"
	// "crypto/md5"
	// "archive/tar"
	// "encoding/json"
	"path/filepath"

	// "github.com/docker/docker/api/types"
	// "github.com/seveirbian/gear/pkg"
	// "github.com/docker/docker/client"
	// "github.com/docker/docker/daemon/graphdriver/overlay2"
	"github.com/sirupsen/logrus"
	// "golang.org/x/net/context"
)

var (
	logger = logrus.WithField("gear", "push")
)

type Pusher struct {
	StorageIP string
	StoragePort string

	GFilesDir string
	FilesToSent map[string]string

	DoNotClean bool
}

func InitPusher(path, ip, port string, doNotClean bool) (*Pusher, error) {
	noClean := false
	if doNotClean == true {
		noClean = true
	}

	return &Pusher {
		StorageIP: ip, 
		StoragePort: port, 
		GFilesDir: path, 
		FilesToSent: map[string]string{}, 
		DoNotClean: noClean, 
	}, nil
}

func (p *Pusher) Push() {
	// 遍历普通文件目录，将所有文件添加到待push的字典中
    err := filepath.Walk(p.GFilesDir, func(path string, f os.FileInfo, err error) error {
    	if f == nil {
    		return err
    	}

    	// 放置将files文件夹也传过去
		pathSlice := strings.SplitAfter(path, p.GFilesDir)
		if pathSlice[1] == "" {
			return nil
		}

    	p.FilesToSent[f.Name()] = path

    	return nil
    })

    if err != nil {
    	logger.Warnf("Fail to walk dir for %v", err)
    }

    // 将字典中所有文件都询问manager，如果该文件已经存在storage中，则将其从字典中删除
    fmt.Println("Querying...")
    toDelete := map[string]string{}
    for cid, path := range p.FilesToSent {
    	resp, err := http.PostForm("http://"+p.StorageIP+":"+p.StoragePort+"/query/"+cid, url.Values{})
    	if err != nil {
    		logger.Warnf("Fail to query cid for %v", err)
    	}

    	if resp.StatusCode == http.StatusOK {
    		toDelete[cid] = path
    	}
    }

    for cid, _ := range toDelete {
    	delete(p.FilesToSent, cid)
    }

    fmt.Println("Uploading...")
    for cid, path := range p.FilesToSent {
    	// 创建表单文件
	    // CreateFormFile 用来创建表单，第一个参数是字段名，第二个参数是文件名
	    buf := new(bytes.Buffer)
	    writer := multipart.NewWriter(buf)
	   	formFile, err := writer.CreateFormFile("file", cid)
	   	if err != nil {
	        logger.Warnf("Fail to create form file failed: %v", err)
	    }
	    // 从文件读取数据，写入表单
	    srcFile, err := os.Open(path)
	    if err != nil {
	        logger.Warnf("Fail to open source file for %v", err)
	    }
	    _, err = io.Copy(formFile, srcFile)
	    if err != nil {
	        logger.Warnf("Fail to write to form file for %v", err)
	    }
	    // 发送表单
	    contentType := writer.FormDataContentType()
	    writer.Close() // 发送之前必须调用Close()以写入结尾行
	    _, err = http.Post("http://"+p.StorageIP+":"+p.StoragePort+"/push/"+cid, contentType, buf)
	    if err != nil {
	        logger.Fatalf("Post failed: %s\n", err)
	    }
	    srcFile.Close()
    }

    fmt.Println("Push OK!")

    if !p.DoNotClean {
    	fmt.Println("Cleaning up dir: ", p.GFilesDir)
    	err = os.RemoveAll(p.GFilesDir)
	    if err != nil {
	    	logger.Warnf("Fail to remove all files under p.GFilesDir for %v", err)
	    }

    	fmt.Println("Clean up OK!")
    }
}

func ParseImage(image string) (imageName string, imageTag string) {
	registryAndImage := strings.Split(image, "/")

	// dockerhub镜像
	if len(registryAndImage) == 1 {
		imageAndTag := strings.Split(image, ":")

		switch len(imageAndTag) {
		case 1: 
			logger.Warn("No image tag provided, use \"latest\"\n")
			imageName = imageAndTag[0]
			imageTag = "latest"
		case 2:
			imageName = imageAndTag[0]
			imageTag = imageAndTag[1]
		}
	}

	// 私有仓库镜像
	if len(registryAndImage) == 2 {
		imageAndTag := strings.Split(registryAndImage[1], ":")

		switch len(imageAndTag) {
		case 1: 
			logger.Warn("No image tag provided, use \"latest\"\n")
			imageName = registryAndImage[0] + "/" + imageAndTag[0]
			imageTag = "latest"
		case 2:
			imageName = registryAndImage[0] + "/" + imageAndTag[0]
			imageTag = imageAndTag[1]
		}
	}

	return
}

























