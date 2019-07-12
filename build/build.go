package build

import (
	"fmt"
	"io"
	"os"
	// "bytes"
	"strings"
	"io/ioutil"
	// "crypto/md5"
	"archive/tar"
	"encoding/json"
	"path/filepath"

	gzip "github.com/klauspost/pgzip"
	"github.com/docker/docker/api/types"
	"github.com/seveirbian/gear/pkg"
	"github.com/docker/docker/client"
	"github.com/docker/docker/daemon/graphdriver/overlay2"
	// "github.com/seveirbian/gear/graphdriver"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var (
	logger = logrus.WithField("gear", "build")
)

type DockerFile struct {
	FROM       string
	ENV        []string
	RUN        []string
	LABEL      map[string]string
	EXPOSE     map[string]struct{} // "80/tcp":{}
	ENTRYPOINT []string
	VOLUME     map[string]struct{}
	WORKDIR    string
	USER       string
	CMD        []string
}

type Builder struct {
	DImageName string
	DImageTag  string

	DOverlayID string

	DImageInfo types.ImageInspect //docker image infomation got by docker inspect

	GImageName string
	GImageTag  string

	Ctx    context.Context
	Client *client.Client

	GearPath           string
	GearBuildPath      string
	RegularFilesPath   string
	IrregularFilesPath string

	IrregularFiles map[string]os.FileInfo
	Dockerfile     DockerFile
}

func InitBuilder(image, suffix string) (*Builder, error) {
	// 1. parse DImage
	dImageName, dImageTag := parseImage(image)

	// 2. init GImage
	gImageName := dImageName + suffix
	gImageTag := dImageTag

	// 3. init docker client which is used to interact with docker daemon
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		logger.Warn("Fail to create docker client...")
		return nil, err
	}

	// 4. get docker image info
	imageInfo, _, err := cli.ImageInspectWithRaw(ctx, dImageName+":"+dImageTag)
	if err != nil {
		logger.Warnf("Fail to inspect image: %s\n", dImageName+":"+dImageTag)
		return nil, err
	}

	// 5. init build path
	gearPath := "/var/lib/gear/"
	_, err = os.Stat(gearPath)
	if err != nil {
		err = os.MkdirAll(gearPath, os.ModePerm)
		if err != nil {
			logger.Warn("Fail to create gearPath...")
			return nil, err
		}
	}
	gearBuildPath := filepath.Join(gearPath, "build")
	_, err = os.Stat(gearBuildPath)
	if err != nil {
		err = os.MkdirAll(gearBuildPath, os.ModePerm)
		if err != nil {
			logger.Warn("Fail to create gearBuildPath...")
			return nil, err
		}
	}
	regularFilesPath := filepath.Join(gearBuildPath, gImageName + ":" + gImageTag, "files")
	_, err = os.Stat(regularFilesPath)
	if err != nil {
		err = os.MkdirAll(regularFilesPath, os.ModePerm)
		if err != nil {
			logger.Warn("Fail to create regularFilesPath...")
			return nil, err
		}
	}
	irregularFilesPath := filepath.Join(gearBuildPath, gImageName + ":" + gImageTag, "build")
	_, err = os.Stat(irregularFilesPath)
	if err != nil {
		err = os.MkdirAll(irregularFilesPath, os.ModePerm)
		if err != nil {
			logger.Warn("Fail to create irregularFilesPath...")
			return nil, err
		}
	}

	// 6. get Dimage's lower layer paths and upper layer path
	var dOverlayID string
	dOverlayID = imageInfo.GraphDriver.Data["UpperDir"]
	// dOverlayID = strings.Split(dOverlayID, "/var/lib/docker/overlay2/")[1]
	dOverlayID = strings.Split(dOverlayID, "/var/lib/docker/overlay2/")[1]
	dOverlayID = strings.Split(dOverlayID, "/diff")[0]

	return &Builder{
		DImageName:         dImageName,
		DImageTag:          dImageTag,
		DOverlayID:         dOverlayID,
		DImageInfo:         imageInfo,
		GImageName:         gImageName,
		GImageTag:          gImageTag,
		Ctx:                ctx,
		Client:             cli,
		GearPath:           gearPath,
		GearBuildPath:      gearBuildPath,
		RegularFilesPath:   regularFilesPath,
		IrregularFilesPath: irregularFilesPath,
	}, nil
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

func (b *Builder) Build(recordedFiles, recordedFileNames []string) error {
	// 1. mount these path and tar irregular files into tmp.tar and
	// copy regular files to GearBuildPath/imageID/common/
	fmt.Println("Tar tmp.tar and copy regular files...")
	err := b.tarAndCopy(recordedFiles, recordedFileNames)
	if err != nil {
		logger.Warn("Fail to walk through layers of this image...")
		return err
	}

	// 2. create dockerfile and put it in GearBuildPath/imageID/build/
	fmt.Println("Creating Dockerfile...")
	err = b.createDocekrfile()
	if err != nil {
		logger.Warn("Fail to create gear dockerfile...")
		return err
	}

	// 3. build gear image by using docker build cmd
	fmt.Println("Building gear image...")
	err = b.buildGearImage()
	if err != nil {
		logger.Warnf("Fail to build gear index image...")
		return err
	}

	return nil
}

func (b *Builder) tarAndCopy(recordedFiles, recordedFileNames []string) error {
	// 1. mount lower layer paths and upper layer path using overlayfs
	driver, err := overlay2.Init("/var/lib/docker/overlay2", []string{}, nil, nil)
	if err != nil {
		logger.WithField("err", err).Warn("Fail to create overlay2 driver...")
		return err
	}

	mountPath, err := driver.Get(b.DOverlayID, "")
	if err != nil {
		logger.WithField("err", err).Warn("Fail to mount overlayfs...")
		return err
	}
	defer driver.Put(b.DOverlayID)

	mergedPath := mountPath.Path()

	// 2. tar irregular files into /var/lib/gear/build/imageID/build/tmp.tar and
	// copy regular files to /var/lib/gear/build/imageID/common/
	tmpFile, err := os.Create(filepath.Join(b.IrregularFilesPath, "tmp.tar"))
	if err != nil {
		logger.Warnf("Fail to create tmp.tar for %v", err)
		return err
	}
	defer tmpFile.Close()

	tw := tar.NewWriter(tmpFile)
	defer tw.Close()

	// 判断是否已经创建了软链接
	_, err = os.Lstat(filepath.Join(mergedPath, "gear-image"))
	if err == nil {
		err := os.Remove(filepath.Join(mergedPath, "gear-image"))
		if err != nil {
			logger.Warnf("Fail to remove gear-image symlink for %v", err)
		}
	}
	err = os.Symlink(b.GImageName+":"+b.GImageTag, filepath.Join(mergedPath, "gear-image"))
	if err != nil {
		fmt.Println(err)
		logger.Warn("Fail to create gear-image file...")
		return err
	}

	err = filepath.Walk(mergedPath, func(path string, f os.FileInfo, err error) error {
		// fail to get file info
		if f == nil {
			return err
		}

		var target string

		// get file's relative path
		var finalPath string
		pathSlice := strings.SplitAfter(path, mergedPath)
		if pathSlice[1] == "" {
			return nil
		}
		for i := 1; i < len(pathSlice); i++ {
			finalPath += pathSlice[i]
		}
		finalPath = strings.TrimPrefix(finalPath, string(filepath.Separator))

		// current file is a regular file
		if f.Mode().IsRegular() {
			src, err := os.Open(path)
			if err != nil {
				logger.Warnf("Fail to open file for %v", err)
				return err
			}
			defer src.Close()

			hashValue := []byte(pkg.HashAFileInMD5(path))

			_, err = os.Lstat(filepath.Join(b.RegularFilesPath, string(hashValue)))
			if err != nil {
				// 创建压缩的普通文件
				dst, err := os.Create(filepath.Join(b.RegularFilesPath, string(hashValue)))
				if err != nil {
					logger.WithField("err", err).Warnf("Fail to create file: %s\n", filepath.Join(b.RegularFilesPath, string(hashValue)))
					return err
				}
				defer dst.Close()

				gw := gzip.NewWriter(dst)
				defer gw.Close()

				srcContent, err := ioutil.ReadAll(src)
				if err != nil {
					logger.Warnf("Fail to read file all for %v", err)
				}

				_, err = gw.Write(srcContent)
				if err != nil {
					logger.Warnf("Fail to write gzip file for %v", err)
				}

				// 修改文件属性
				err = os.Chmod(filepath.Join(b.RegularFilesPath, string(hashValue)), f.Mode().Perm())
				if err != nil {
					logger.Warnf("Fail to chmod for %v", err)
				}
			}


			// 将普通文件的内容替换成哈希值
			hd, err := tar.FileInfoHeader(f, target)
			if err != nil {
				logger.Warn("Fail to get file head...")
				return err
			}

			hd.Name = finalPath

			hd.Size = int64(len(hashValue))

			// write file header info
			err = tw.WriteHeader(hd)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write header info")
				return err
			}

			_, err = tw.Write(hashValue)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write content...")
				return err
			}

			return nil
		}

		// current file is a symlink
		if f.Mode()&os.ModeSymlink != 0 {
			target, err = os.Readlink(path)
			if err != nil {
				logger.Warn("Fail to read symlink target...")
				return err
			}
		}

		hd, err := tar.FileInfoHeader(f, target)
		if err != nil {
			logger.Warn("Fail to get file head...")
			return err
		}

		hd.Name = finalPath

		// write file header info
		err = tw.WriteHeader(hd)
		if err != nil {
			logger.WithField("err", err).Warn("Fail to write header info")
			return err
		}

		return nil
	})

	if err != nil {
		logger.Warn("Fail to walk layers of image...")
		return err
	}

	if recordedFiles != nil {
		content := ""

		if len(recordedFiles) != len(recordedFileNames) {
			logger.Warnf("Something went error that len(recordedFiles) != len(recordedFileNames)")
		} else {
			for i := 0; i < len(recordedFiles) - 1; i++ {
				content = content+recordedFileNames[i]+" "+recordedFiles[i]+"\n"
			}
			content = content+recordedFileNames[len(recordedFiles)-1]+" "+recordedFiles[len(recordedFiles)-1]
		}

		_, err := os.Create(filepath.Join(mergedPath, "RecordFiles"))
		if err != nil {
			logger.Warnf("Fail to create RecordFiles for %v", err)
		}

		f, err := os.Lstat(filepath.Join(mergedPath, "RecordFiles"))
		if err != nil {
			logger.Warnf("Fail to Lstat RecordFiles for %v", err)
		}

		hd, err := tar.FileInfoHeader(f, "")
		if err != nil {
			logger.Warnf("Fail to create tar header for %v", err)
		}

		hd.Name = "/RecordFiles"

		hd.Size = int64(len(content))

		// write file header info
		err = tw.WriteHeader(hd)
		if err != nil {
			logger.WithField("err", err).Warn("Fail to write header info")
			return err
		}

		_, err = tw.Write([]byte(content))
		if err != nil {
			logger.WithField("err", err).Warn("Fail to write content...")
			return err
		}
	}

	return nil
}

func (b *Builder) createDocekrfile() error {
	// 1. fill b.Dockerfile struct
	b.Dockerfile.FROM = "scratch"
	b.Dockerfile.ENV = b.DImageInfo.Config.Env
	b.Dockerfile.LABEL = b.DImageInfo.Config.Labels
	b.Dockerfile.VOLUME = b.DImageInfo.Config.Volumes
	b.Dockerfile.WORKDIR = b.DImageInfo.Config.WorkingDir
	b.Dockerfile.USER = b.DImageInfo.Config.User

	b.Dockerfile.EXPOSE = map[string]struct{}{}
	exposedPorts := b.DImageInfo.Config.ExposedPorts
	for key, value := range exposedPorts {
		b.Dockerfile.EXPOSE[string(key)] = value
	}

	entryPoints := b.DImageInfo.Config.Entrypoint
	for _, value := range entryPoints {
		b.Dockerfile.ENTRYPOINT = append(b.Dockerfile.ENTRYPOINT, string(value))
	}

	cmds := b.DImageInfo.Config.Cmd
	for _, value := range cmds {
		b.Dockerfile.CMD = append(b.Dockerfile.CMD, string(value))
	}

	// 2. transform b.Dockerfile to []bytes
	var dockerfile string

	dockerfile = dockerfile + "FROM "
	dockerfile = dockerfile + b.Dockerfile.FROM
	dockerfile = dockerfile + "\n"

	for _, env := range b.Dockerfile.ENV {
		dockerfile = dockerfile + "ENV "
		envSlices := strings.Split(env, "=")
		dockerfile = dockerfile + envSlices[0] + " "
		dockerfile = dockerfile + "\"" + envSlices[1] + "\""
		dockerfile = dockerfile + "\n"
	}

	for key, value := range b.Dockerfile.LABEL {
		dockerfile = dockerfile + "LABEL "
		dockerfile = dockerfile + key + "=\"" + value + "\""
		dockerfile = dockerfile + "\n"
	}

	for key, _ := range b.Dockerfile.VOLUME {
		dockerfile = dockerfile + "VOLUME "
		dockerfile = dockerfile + key
		dockerfile = dockerfile + "\n"
	}

	if b.Dockerfile.WORKDIR != "" {
		dockerfile = dockerfile + "WORKDIR "
		dockerfile = dockerfile + b.Dockerfile.WORKDIR
		dockerfile = dockerfile + "\n"
	}

	for key, _ := range b.Dockerfile.EXPOSE {
		dockerfile = dockerfile + "EXPOSE "
		dockerfile = dockerfile + strings.Split(key, "/")[0]
		dockerfile = dockerfile + "\n"
	}

	if len(b.Dockerfile.ENTRYPOINT) != 0 {
		dockerfile = dockerfile + "ENTRYPOINT ["
		for _, entry := range b.Dockerfile.ENTRYPOINT {
			dockerfile = dockerfile + "\"" + entry + "\", "
		}
		dockerfile = strings.TrimSuffix(dockerfile, ", ")
		dockerfile = dockerfile + "]\n"
	}

	dockerfile = dockerfile + "ADD "
	dockerfile = dockerfile + "tmp.tar /"
	dockerfile = dockerfile + "\n"

	if b.Dockerfile.USER != "" {
		dockerfile = dockerfile + "USER "
		dockerfile = dockerfile + b.Dockerfile.USER
		dockerfile = dockerfile + "\n"
	}

	if len(b.Dockerfile.CMD) != 0 {
		dockerfile = dockerfile + "CMD ["
		for _, cmd := range b.Dockerfile.CMD {
			dockerfile = dockerfile + "\"" + cmd + "\", "
		}
		dockerfile = strings.TrimSuffix(dockerfile, ", ")
		dockerfile = dockerfile + "]\n"
	}

	// 3. write to dockerfile in gearBuildPath
	f, err := os.Create(filepath.Join(b.IrregularFilesPath, "Dockerfile"))
	if err != nil {
		logger.Warn("Fail to create Dockerfile...")
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(dockerfile))
	if err != nil {
		logger.Warn("Fail to write to Dockerfile...")
		return err
	}

	return nil
}

func (b *Builder) buildGearImage() error {
	// 1. create a tarball which contains everything including Dockerfile
	var files = []string{
		filepath.Join(b.IrregularFilesPath, "tmp.tar"),
		filepath.Join(b.IrregularFilesPath, "Dockerfile"),
	}

	var buildTarPath = filepath.Join(b.IrregularFilesPath, "build.tar")
	err := archive(files, buildTarPath)
	if err != nil {
		logger.Warn("Fail to create build.tar...")
		return err
	}

	// 2. open build.tar
	buildTar, err := os.Open(buildTarPath)
	if err != nil {
		logger.Warn("Fail to open build.tar...")
		return err
	}

	// 3. init image build options
	opts := types.ImageBuildOptions{
		NoCache: true, 
		Context: buildTar,
		Tags:    []string{b.GImageName + ":" + b.GImageTag},
	}

	// 4. start to build
	buildResp, err := b.Client.ImageBuild(b.Ctx, buildTar, opts)
	if err != nil {
		logger.WithField("err", err).Warn("Fail to build...")
		return err
	}
	defer buildResp.Body.Close()

	// following two raws are important, cannot be removed
	type Message struct {
		Stream string
	}
	decoder := json.NewDecoder(buildResp.Body)
	for decoder.More() {
		var m Message
		err := decoder.Decode(&m)
		if err != nil {
			logger.WithField("err", err).Warn("Fail decode build response...")
			return err
		}
		fmt.Print(m.Stream)
	}

	return nil
}

func archive(files []string, archivePath string) error {
	// 1. create a file, which will store the tar data
	tarFile, err := os.Create(archivePath)
	if err != nil {
		logger.WithField("err", err).Warn("Fail to create archive target file...")
		return err
	}
	defer tarFile.Close()

	// 2. create a tar writer on tarFile
	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	for _, file := range files {
		fInfo, err := os.Lstat(file)
		if err != nil {
			logger.WithField("err", err).Warn("Fail to get file's info...")
			return err
		}

		mode := fInfo.Mode()
		var target string

		// if this file is a regular file
		if mode.IsRegular() {
			// get file header info
			hd, err := tar.FileInfoHeader(fInfo, target)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to get file header...")
				return err
			}
			// write file header info
			err = tw.WriteHeader(hd)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write file header...")
				return err
			}
			// open file
			f, err := os.Open(file)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to open file...")
				return err
			}
			defer f.Close()
			// copy to tar file
			_, err = io.Copy(tw, f)
			if err != nil {
				logger.WithField("err", err).Warn("Fail to write tar file...")
				return err
			}
		}
	}

	return nil
}
