package monitor

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
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

	fmt.Println(image)
	fmt.Println(files)

	return c.NoContent(http.StatusOK)
}