package fs

import (
	"errors"
	"strings"
	"path/filepath"
)

var (

)

func ValidatePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("Not a absolute path...")
	}

	if !strings.HasSuffix(path, "/") {
		return path+"/", nil
	}

	return path, nil
}