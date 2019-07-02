package monitor

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/seveirbian/gear/build"
	// gearTypes "github.com/seveirbian/gear/types"
)

var (

)

func handleEvent(c echo.Context) error {
	// 1. 获取镜像名
	image := c.Param("IMAGE")
	// 2. 获取文件
	values, err := c.FormParams()
	if err != nil {
		logger.Warnf("Fail to parse files for %v", err)
	}

	files := values["files"]

	builder, err := build.InitBuilder(image)
	if err != nil {
		logger.Fatal("Fail to init a builder to build gear image...")
	}
	err = builder.Build(files)
	if err != nil {
		logger.Fatal("Fail to build gear image...")
	}

	fmt.Println(image)
	fmt.Println(files)

	return c.NoContent(http.StatusOK)
}