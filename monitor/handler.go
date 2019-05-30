package monitor

import (
	"net/http"

	"github.com/labstack/echo"
)

var (

)

func handleEvent(c echo.Context) error {
	path := c.FormValue("path")

	AccessedFiles = append(AccessedFiles, path)

	return c.NoContent(http.StatusOK)
}