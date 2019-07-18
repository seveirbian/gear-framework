package monitor

import (
	"os"
	"fmt"
	"time"
	// "path"
	"os/exec"
	"strings"
	"net/http"
	"strconv"
	"io/ioutil"
	"archive/tar"
	"math/rand"
	"path/filepath"
	"encoding/json"

	"github.com/labstack/echo"
	// gzip "github.com/klauspost/pgzip"
	"github.com/seveirbian/gear/build"
	// "github.com/docker/docker/api/types"
	// gearTypes "github.com/seveirbian/gear/types"
)

var (

)

func handleEvent(c echo.Context) error {
	values, err := c.FormParams()
	if err != nil {
		logger.Warnf("Fail to parse files for %v", err)
	}

	c.NoContent(http.StatusOK)

	// 1. 获取镜像名
	image := values["image"][0] // 202.114.10.146:9999/tomcat-gear:8
	imageRepo, imageTag := parseImage(image)
	image = strings.TrimSuffix(imageRepo, "-gear") + ":" + imageTag

	// 2. 获取文件
	files := values["files"]
	names := values["filenames"]

	fmt.Println("image name: ", image)

	// 3. 检查是否已经构建过
	if !check(strings.TrimSuffix(imageRepo, "-gear") + "-gearmd", imageTag) {
		// 3. 构建包含预取文件的新gear镜像
		builder, err := build.InitBuilder(image, "-gearmd")
		if err != nil {
			logger.Fatal("Fail to init a builder to build gear image...")
		}
		err = builder.Build(files, names)
		if err != nil {
			logger.Fatal("Fail to build gear image for %v", err)
		}

		// 4. push -gearmd镜像
		mdImage := strings.TrimSuffix(imageRepo, "-gear") + "-gearmd" + ":" + imageTag
	    cName := "docker"
	    cArgs := []string{"push", mdImage}
	    cCmd := exec.Command(cName, cArgs...)
	    if err := cCmd.Run(); err != nil {
	        fmt.Fprintln(os.Stderr, err)
	        os.Exit(1)
	    }
	    fmt.Println("push", mdImage, "done!")

		fmt.Println("Push ok!")
	} else {
		fmt.Println("Have built it!")
	}

	// mnt.Client.ImageRemove(mnt.Ctx, )

	return nil
}

func check(repo, tag string) bool {
	repo = strings.TrimPrefix(repo, mnt.RegistryIp+":"+mnt.RegistryPort+"/")

	resp, err := http.Get("http://"+mnt.RegistryIp+":"+mnt.RegistryPort+"/v2/"+repo+"/tags/list")
	if err != nil {
		logger.Warnf("Fail to get tags of %s", repo)
	}
	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("Fail to read...")
	}
	type Tags struct {
		Name string `json:"name"`
		Tags []string `json:"tags"`
	}
	var tags Tags
	json.Unmarshal(rs, &tags)

	// fmt.Println(tags)

	for _, btTag := range tags.Tags {
		// fmt.Println(repo, "  ", tag, "  ", btTag)
		if btTag == tag {
			return true
		}
	}

	return false
}

func createGzip(files []string, gzipPath string, image string) error {
	_, err := os.Lstat(filepath.Join(gzipPath, image))
	if err != nil {
		// 没有预先创建好的压缩包
		rand.Seed(time.Now().Unix())
		tmpFileName := strconv.Itoa(rand.Int())

		fmt.Println(tmpFileName)

		imageGzip, err := os.Create(filepath.Join(gzipPath, tmpFileName))
		if err != nil {
			logger.Warnf("Fail to create file for %v", err)
		}
		defer os.Remove(filepath.Join(gzipPath, tmpFileName))

		tw := tar.NewWriter(imageGzip)

		for _, file := range files {
			if file == "" {
				continue
			}

			f, err := os.Stat(filepath.Join(GearStoragePath, file))
			if err != nil {
				logger.Warnf("Fail to stat file for %v", err)
				continue
			}

			hd, err := tar.FileInfoHeader(f, "")
			if err != nil {
				logger.Warn("Fail to get file head...")
				return err
			}

			err = tw.WriteHeader(hd)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write header info")
				return err
			}

			b, err := ioutil.ReadFile(filepath.Join(GearStoragePath, file))
			if err != nil {
				logger.Warnf("Fail to read file for %v", err)
			}

			_, err = tw.Write(b)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write content...")
				return err
			}
		}

		tw.Close()
		imageGzip.Close()

		// 开始压缩
		// _, err = os.Lstat(path.Dir(filepath.Join(gzipPath, image)))
		// if err != nil {
		// 	err = os.MkdirAll(path.Dir(filepath.Join(gzipPath, image)), os.ModePerm)
		// 	if err != nil {
		// 		logger.Warnf("Fail to create parent dir for %v", err)
		// 	}
		// }

		// gzipFile, err := os.Create(filepath.Join(gzipPath, image))
		// if err != nil {
		// 	logger.Warnf("Fail to create gzip file for %v", err)
		// }

		// gw := gzip.NewWriter(gzipFile)

		// tarContent, err := ioutil.ReadFile(filepath.Join(gzipPath, tmpFileName))
		// if err != nil {
		// 	logger.Warnf("Fail to read tmp file for %v", err)
		// }

		// _, err = gw.Write(tarContent)
		// if err != nil {
		// 	logger.Warnf("Fail to write gzip file for %v", err)
		// }

		// gw.Close()
		// gzipFile.Close()
	}

	return nil
}

func parseImage(image string) (imageName string, imageTag string) {
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



