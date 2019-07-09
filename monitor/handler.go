package monitor

import (
	"os"
	"fmt"
	"time"
	"path"
	"os/exec"
	"strings"
	"net/http"
	"strconv"
	"io/ioutil"
	"archive/tar"
	"math/rand"
	"path/filepath"

	"github.com/labstack/echo"
	gzip "github.com/klauspost/pgzip"
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

	// 1. 获取镜像名
	image := values["image"][0] // 202.114.10.146:9999/tomcat-gear:8
	imageSlices := strings.Split(image, ":")
	imageRepo := ""
	for i := 0; i < len(imageSlices) - 2; i++ {
		imageRepo = imageRepo + imageSlices[i] + ":"
	}
	imageRepo += strings.TrimSuffix(imageSlices[len(imageSlices) - 2], "-gear")
	image = imageRepo + ":" + imageSlices[len(imageSlices) - 1]


	// 2. 获取文件
	files := values["files"]

	fmt.Println("image name: ", image)
	fmt.Println("files: ", files)

	// 1. 创建镜像的压缩文件
	err = createGzip(files, GearGzipPath, image)
	if err != nil {
		logger.Warnf("Fail to create image gizp file for %v", err)
	}

	// 2. 构建包含预取文件的新gear镜像
	builder, err := build.InitBuilder(image, "-gearmd")
	if err != nil {
		logger.Fatal("Fail to init a builder to build gear image...")
	}
	err = builder.Build(files)
	if err != nil {
		logger.Fatal("Fail to build gear image...")
	}

	slices := strings.Split(image, ":")
	repo := ""
	for i := 0; i < len(slices) - 2; i++ {
		repo = repo + slices[i] + ":"
	}
	repo = repo + slices[len(slices)-2]
	tag := slices[len(slices)-1]

	fmt.Println(repo)
	fmt.Println(tag)

    cName := "docker"
    cArgs := []string{"push", repo+"-gear"+":"+tag}
    cCmd := exec.Command(cName, cArgs...)
    if err := cCmd.Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Println("done!")

	fmt.Println(image)
	fmt.Println(values["id"])
	fmt.Println(files)
	fmt.Println("Push ok!")

	return c.NoContent(http.StatusOK)
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
		_, err = os.Lstat(path.Dir(filepath.Join(gzipPath, image)))
		if err != nil {
			err = os.MkdirAll(path.Dir(filepath.Join(gzipPath, image)), os.ModePerm)
			if err != nil {
				logger.Warnf("Fail to create parent dir for %v", err)
			}
		}

		gzipFile, err := os.Create(filepath.Join(gzipPath, image))
		if err != nil {
			logger.Warnf("Fail to create gzip file for %v", err)
		}

		gw := gzip.NewWriter(gzipFile)

		tarContent, err := ioutil.ReadFile(filepath.Join(gzipPath, tmpFileName))
		if err != nil {
			logger.Warnf("Fail to read tmp file for %v", err)
		}

		_, err = gw.Write(tarContent)
		if err != nil {
			logger.Warnf("Fail to write gzip file for %v", err)
		}

		gw.Close()
		gzipFile.Close()
	}

	return nil
}





