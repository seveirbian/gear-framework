package monitor

import (
	// "fmt"
	// "strings"
	"net/http"

	"github.com/labstack/echo"
	// "github.com/seveirbian/gear/build"
	// "github.com/docker/docker/api/types"
	// gearTypes "github.com/seveirbian/gear/types"
)

var (

)

func handleEvent(c echo.Context) error {
	// 1. 获取镜像名
	image := c.Param("IMAGE")
	// 2. 获取文件
	values, err := c.FormParams()
	// if err != nil {
	// 	logger.Warnf("Fail to parse files for %v", err)
	// }

	// files := values["files"]

	// builder, err := build.InitBuilder(image)
	// if err != nil {
	// 	logger.Fatal("Fail to init a builder to build gear image...")
	// }
	// err = builder.Build(files)
	// if err != nil {
	// 	logger.Fatal("Fail to build gear image...")
	// }

	// slices := strings.Split(image, ":")
	// repo := ""
	// for i := 0; i < len(slices) - 1; i++ {
	// 	repo += slices[i]
	// }
	// tag := slices[len(slices)-1]

	// res, err := mnt.Client.ImagePush(mnt.Ctx, repo+"-gear"+":"+tag, types.ImagePushOptions{RegistryAuth: "123"})
 //    if err != nil {
 //    	logger.Warnf("Fail to push images for %v", err)
 //    }
 //    defer res.Close()

	// fmt.Println(image)
	// fmt.Println(values["id"])
	// fmt.Println(files)
	// fmt.Println("Push ok!")

	return c.NoContent(http.StatusOK)
}