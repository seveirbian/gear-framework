package monitor

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

var (

)

func handleEvent(c echo.Context) error {
	path := c.FormValue("path")
	hash := c.FormValue("hash")

	fmt.Println(path, hash)

	AccessedFiles = append(AccessedFiles, path)

	return c.NoContent(http.StatusOK)
}