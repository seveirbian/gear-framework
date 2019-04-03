package client

import (
	"io"
	"os"
	"fmt"
	"strings"
	"syscall"
	"strconv"
	"net/url"
	"net/http"
	"encoding/json"
	"path/filepath"
	"github.com/labstack/echo"
	"github.com/seveirbian/gear/pkg"
	"github.com/seveirbian/gear/types"
)

var ()

func handleInfo(c echo.Context) error {
	resp := strconv.FormatUint(cli.Self.ID, 10)+":"+cli.Self.IP+":"+cli.Self.Port

	return c.String(http.StatusOK, resp)
}

func handleDownload(c echo.Context) error {
	// 1. 获取请求的OID
	cid := c.Param("CID")

	// 2. 检测本地nfs是否还在成功运行
 
	// 3. TODO：否则，怎么办呢

	// 4. 查看本地缓存是否存在
	_, err := os.Stat(filepath.Join("/var/lib/gear/public", cid))

	// 不存在
	if err != nil {
		// // 返回nfs目录下的文件
		// err = c.Attachment(filepath.Join("/var/lib/gear/nfs", cid), cid)
		// if err != nil {
		// 	fmt.Println(err)
		// 	logger.Fatal("Fail to return nfs file...")
		// }
		// 将文件拷贝到共享目录下
		dst, err := os.Create(filepath.Join("/var/lib/gear/public", cid))
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to create sharing file...")
		}
		defer dst.Close()
		// 上独占文件锁
		err = syscall.Flock(int(dst.Fd()), syscall.LOCK_EX)
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to lock file in an ex way...")
		}
		src, err := os.Open(filepath.Join("/var/lib/gear/nfs", cid))
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to open nfs file...")
		}
		defer src.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to copy file...")
		}
		// 解锁
		err = syscall.Flock(int(dst.Fd()), syscall.LOCK_UN)
		if err != nil {
			fmt.Println(err)
			logger.Fatal("Fail to unlock file in an ex way...")
		}
	}

	// 返回本地文件
	f, err := os.Open(filepath.Join("/var/lib/gear/public", cid))
	if err != nil {
		fmt.Println(err)
		logger.Fatalf("Fail to open file: %s\n", filepath.Join("/var/lib/gear/public", cid))
	}
	defer f.Close()
	// 上共享文件锁
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Fail to lock file in a sharing way...")
	}
	err = c.Attachment(filepath.Join("/var/lib/gear/public", cid), cid)
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Fail to return file...")
	}
	// 解锁
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Fail to unlock file in a sharing way...")
	}

	return nil
}

func handleGet(c echo.Context) error {
	imageName := c.Param("IMAGENAME")
	tag := c.Param("TAG")
	path := c.FormValue("PATH")

	// 1. 获取OID
	oid := pkg.Hash(imageName+":"+tag+path)

	// 2. 查询etcd中OID对应CID
	resp, err := http.Get("http://"+cli.Etcd.IP+":"+cli.Etcd.Port+"/v2/keys/"+strconv.FormatUint(oid, 10))
	if err != nil {
		logger.Fatal("Fail to get cid...")
	}

	// 若没有找到，则返回not found
	if resp.StatusCode == http.StatusNotFound {
		return c.NoContent(http.StatusNotFound)
	}

	// 找到了则接受json
	var response types.Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Fail to decode resp...")
	}

	cid := response.EtcdNode.Value

	// 3. 查看当前文件是否已经存在在private文件夹中
	_, err = os.Stat(filepath.Join("/var/lib/gear/private", cid))
	if err == nil {
		fmt.Println("already have it")
		err = os.Link(filepath.Join("/var/lib/gear/private", cid), filepath.Join("/var/lib/gear/images", imageName+":"+tag, path))
		if err != nil {
			logger.Fatal("Fail to create hard link...")
		}

		return c.NoContent(http.StatusOK)
	}

	// 4. 搜索所有集群节点，找到最优节点获取文件
	cidInUint64, _ := strconv.ParseUint(cid, 10, 64)
	var distance uint64 = cidInUint64 ^ cli.Self.ID
	candidata := cli.Self

	cli.NodesMu.RLock()
	for cliId, client := range(cli.Nodes) {
		if distance > (cidInUint64 ^ cliId) {
			distance = cidInUint64 ^ cliId
			candidata = client
		}
	}
	cli.NodesMu.RUnlock()

	// 5. 下载到/var/lib/gear/private
	resp, err = http.PostForm("http://"+candidata.IP+":"+candidata.Port+"/get/"+cid, url.Values{})
	if err != nil {
		logger.Fatal("Fail to get file...")
	}

	// 下载文件
	// filename := (strings.Split(resp.Header.Get("Content-Disposition"), "\"")[1])
	f, err := os.Create(filepath.Join("/var/lib/gear/private", cid))
	if err != nil {
		logger.Fatal(err)
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		logger.Fatal(err)
	}

	// 6. 创建硬连接到imageID（IMAGENAME:TAG）目录下
	err = os.Link(filepath.Join("/var/lib/gear/private", cid), filepath.Join("/var/lib/gear/images", imageName+":"+tag, path))
	if err != nil {
		logger.Fatal("Fail to create hard link...")
	}

	return c.NoContent(http.StatusOK)
}

func handleUpload(c echo.Context) error {
	// 1. get imageName and tag
	imageName := c.Param("IMAGENAME")
	tag := c.Param("TAG")
	// 2. get gear images' path
	gearImagesPath := "/var/lib/gear/images"
	imagePath := filepath.Join(gearImagesPath, imageName+":"+tag)

	err := filepath.Walk(imagePath, func(path string, f os.FileInfo, err error) error {
		// fail to get file info
		if f == nil {
			return err
		}

		// get file's relative path
		pathSlice := strings.Split(path, gearImagesPath)
		if pathSlice[1] == "" {
			return nil
		}
		relativePath := pathSlice[1]

		// current file is a regular file
		if f.Mode().IsRegular() {
			// 1. 计算每个普通文件的cid以及对应的oid
			oid := pkg.Hash(imageName+":"+tag+relativePath)
			cid := pkg.HashAFileInMD5(path)
			// fmt.Printf("%T: %d\n", oid, oid)
			// fmt.Printf("%T: %s\n", cid, cid)

			// 2. 将oid和cid对添加到etcd中
			request, err := http.NewRequest("PUT", "http://"+cli.Etcd.IP+":"+cli.Etcd.Port+"/v2/keys/"+strconv.FormatUint(oid, 10)+"?value="+cid, nil)
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				logger.Fatal("Fail to do request...")
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				fmt.Println("Add oid:cid to etcd!")
			} else if resp.StatusCode == http.StatusOK {
				fmt.Println("Update oid:cid to etcd!")
			} else {
				logger.Fatal("Fail to store oid:cid to etcd...")
			}

			// 3. 将文件以cid的名字复制到nfs目录中
			_, err = os.Stat(filepath.Join("/var/lib/gear/nfs", cid))
			if err != nil {
				// 当前cid文件并没有添加到nfs中，将其复制过去
				dst, err := os.Create(filepath.Join("/var/lib/gear/nfs", cid))
				if err != nil {
					logger.Fatal("Fail to create dst file...")
				}
				defer dst.Close()
				src, err := os.Open(path)
				if err != nil {
					logger.Fatal("Fail to open src file...")
				}
				defer src.Close()
				_, err = io.Copy(dst, src)
				if err != nil {
					logger.Fatal("Fail to copy from src to dst...")
				}
			}

			// 该cid文件已经在nfs中存在了，直接返回即可
		}

		return nil
	})

	if err != nil {
		logger.Fatal("Fail to walk image's dir...")
	}

	fmt.Printf("Upload %s:%s OK!\n", imageName, tag)
	return c.NoContent(http.StatusOK)
}













