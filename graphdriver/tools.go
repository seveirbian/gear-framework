package graphdriver

import (
    "path/filepath"
    "os"
    "strings"
    "io/ioutil"
    "io"
    "github.com/seveirbian/gear/pkg"
)

var (
    
)

func extractAndSaveRegularFiles(source string, target string) (err error) {
	err = filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
			if f == nil {
	    		return err
	    	}

	    	// 防止本目录也处理了
			pathSlice := strings.SplitAfter(path, source)
			if pathSlice[1] == "" {
				return nil
			}

			if f.Mode().IsRegular() {
				hashValue := []byte(pkg.HashAFileInMD5(path))

				src, err := os.Open(path)
				if err != nil {
					logger.Warnf("Fail to open file: %s\n", path)
					return err
				}
				defer src.Close()

				dst, err := os.Create(filepath.Join(target, string(hashValue)))
				if err != nil {
					logger.Warnf("Fail to create file: %s\n", filepath.Join(target, string(hashValue)))
					return err
				}
				defer dst.Close()

				// 拷贝文件内容
				_, err = io.Copy(dst, src)
				if err != nil {
					logger.Warn("Fail to copy file...")
					return err
				}

				err = ioutil.WriteFile(path, hashValue, f.Mode().Perm())
				if err != nil {
					logger.Warnf("Fail to write file for %v", err)
				}

				_, err = src.WriteAt(hashValue, 0)
				if err != nil {
					logger.Warnf("Fail to write for %v\n", err)
				}
			}

			return nil
		})

	return 
}