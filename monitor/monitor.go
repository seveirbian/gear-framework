package monitor

import (
	"os"
	"os/exec"
	"io/ioutil"
	"fmt"
	"sync"
	"time"
	"strings"
	"path/filepath"
	"net/http"
	"encoding/json"

	"golang.org/x/net/context"
	"github.com/sirupsen/logrus"
	"github.com/seveirbian/gear/build"
	"github.com/docker/docker/client"
	gearTypes "github.com/seveirbian/gear/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/seveirbian/gear/push"
	"github.com/labstack/echo"
	"github.com/seveirbian/gear/pkg"
	// "github.com/docker/docker/daemon/graphdriver/overlay2"
	// "github.com/fsnotify/fsnotify"
)

var (
	logger = logrus.WithField("gear", "monitor")
	maxTime = time.Duration(time.Second*10)
)

var mnt Monitor = Monitor{}

var AccessedFiles []string

var (
	GearPath             = "/var/lib/gear/"
	GearBuildPath        = filepath.Join(GearPath, "build")
	GearGzipPath         = filepath.Join(GearPath, "gzip")
	GearStoragePath      = filepath.Join(GearPath, "storage")
)

type Monitor struct {
	MonitorIp string
	MonitorPort string

	RegistryIp string
	RegistryPort string

	ManagerIp string
	ManagerPort string

	Server *echo.Echo

	Ctx    context.Context
	Client *client.Client

	HMutex sync.Mutex
	HaveBeenBuild map[string][]string

	TMutex sync.RWMutex
	ToBeBuild map[string][]string

	NoCleanUp bool
}

func InitMonitor(registry string, managerIp, managerPort string, noCleanUp bool) (*Monitor, error) {
	ip, port := parseRegistry(registry)

	// 创建cli用来和dockerd交互
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		logger.Warn("Fail to create docker client...")
		return nil, err
	}

	// 创建服务器
	e := echo.New()
	e.POST("/event", handleEvent)

	// 获取本地ip
	mIp := pkg.GetSelfIp()
	mPort := "2021"

	mnt.MonitorIp = mIp
	mnt.MonitorPort = mPort
	mnt.RegistryIp = ip
	mnt.RegistryPort = port
	mnt.Server = e
	mnt.ManagerIp = managerIp
	mnt.ManagerPort = managerPort
	mnt.Ctx = ctx
	mnt.Client = cli
	mnt.NoCleanUp = noCleanUp

	return &mnt, nil
}

func (m *Monitor) Monitor() error {
	// 启动服务器
	go m.Server.Start(":" + m.MonitorPort)

	// 获取待处理的镜像列表
	for {
		m.getRepositories()
		m.build()
		time.Sleep(maxTime)
	}

	fmt.Println(m.HaveBeenBuild)
	fmt.Println(m.ToBeBuild)

	return nil
}

func (m *Monitor) getRepositories() error {
	// 1. 获取registry中所有repositories
	resp, err := http.Get("http://"+m.RegistryIp+":"+m.RegistryPort+"/v2/_catalog")
	if err != nil {
		logger.Warnf("Fail to query repositories...")
	}
	defer resp.Body.Close()

	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("Fail to read resp.Body...")
	}

	type Repositories struct {
		Repos []string `json:"repositories"`
	}
	var repos Repositories
	
	json.Unmarshal(rs, &repos)
	
	haveBeenBuild := map[string][]string{}

	// 构建havebeenbuild字典
	for _, repo := range repos.Repos {
		if strings.HasSuffix(repo, "-gear") {
			resp, err := http.Get("http://"+m.RegistryIp+":"+m.RegistryPort+"/v2/"+repo+"/tags/list")
			if err != nil {
				logger.Warnf("Fail to get tags of %s", repo)
			}
			rs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Warnf("Fail to read...")
			}
			type Tags struct {
				Name string `json:"name"`
				Tags []string `json:"tags"`
			}
			var tags Tags
			json.Unmarshal(rs, &tags)
			haveBeenBuild[repo] = []string{}
			for _, tag := range tags.Tags {
				haveBeenBuild[repo] = append(haveBeenBuild[repo], tag)
			}
		}
	}

	toBeBuild := map[string][]string{}
	for _, repo := range repos.Repos {
		// 不要再对-gearmd镜像构建啦
		if strings.HasSuffix(repo, "-gearmd") {
			continue
		}

		if !strings.HasSuffix(repo, "-gear") {
			resp, err := http.Get("http://"+m.RegistryIp+":"+m.RegistryPort+"/v2/"+repo+"/tags/list")
			if err != nil {
				logger.Warnf("Fail to get tags of %s", repo)
			}
			rs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Warnf("Fail to read...")
			}
			type Tags struct {
				Name string `json:"name"`
				Tags []string `json:"tags"`
			}
			var tags Tags
			json.Unmarshal(rs, &tags)

			toBeBuild[repo] = []string{}
			// 该镜像的所有tag都没有被处理过
			if _, ok := haveBeenBuild[repo+"-gear"]; !ok {	
				for _, tag := range tags.Tags {
					toBeBuild[repo] = append(toBeBuild[repo], tag)
				}
			} else {
				// 该镜像的部分tag被处理过
				for _, tag := range tags.Tags {
					if !exist(haveBeenBuild[repo+"-gear"], tag) {
						toBeBuild[repo] = append(toBeBuild[repo], tag)
					}
				}
			}
		}
	}


	// 更新m.havebeenbuild和m.tobebuild
	m.HMutex.Lock()
	m.HaveBeenBuild = haveBeenBuild
	m.HMutex.Unlock()

	m.TMutex.Lock()
	m.ToBeBuild = toBeBuild
	m.TMutex.Unlock()

	return nil
}

func exist(set []string, element string) bool {
	for _, ele := range set {
		if ele == element {
			return true
		}
	}
	return false
}

func (m *Monitor) build() {
	m.TMutex.Lock()
	for repository, tags := range m.ToBeBuild {
		for _, tag := range tags {
			t := time.Now()
			err := m.do_build(gearTypes.Image{Repository: repository, Tag: tag})
			if err != nil {
				logger.Warnf("Fail to build %s:%s", repository, tag)
			}
			fmt.Println("build time: ", time.Since(t))
		}
	}
	m.TMutex.Unlock()
}

func (m *Monitor) do_build(image gearTypes.Image) error {
	// 1. 下载待处理镜像
	fmt.Printf("Pulling %s:%s/%s:%s\n", m.RegistryIp, m.RegistryPort, image.Repository, image.Tag)
	out, err := m.Client.ImagePull(m.Ctx, m.RegistryIp+":"+m.RegistryPort+"/"+image.Repository+":"+image.Tag, types.ImagePullOptions{})
	if err != nil {
		logger.Warnf("Fail to pull the image")
	}
	defer out.Close()
	decoder := json.NewDecoder(out)
	for decoder.More() {
		var retMessage jsonmessage.JSONMessage
		err := decoder.Decode(&retMessage)
		if err != nil {
			logger.Warnf("Fail decode pull response...")
			return err
		}
		// fmt.Printf("%s: %s\r\n", retMessage.Status, retMessage.Progress)
	}
	fmt.Println("Pull OK!")

	// 2. 调用build命令，构建gear镜像
	builder, err := build.InitBuilder(m.RegistryIp+":"+m.RegistryPort+"/"+image.Repository+":"+image.Tag, "-gear")
	if err != nil {
		logrus.Fatal("Fail to init a builder to build gear image...")
	}

	err = builder.Build(nil, nil)
	if err != nil {
		logrus.Fatal("Fail to build gear image...")
	}

	// 3. 将gear镜像push到镜像仓库，并将备用文件存储到存储中
	fmt.Printf("Pushing %s:%s/%s:%s\n", m.RegistryIp, m.RegistryPort, image.Repository+"-gear", image.Tag)
    gFIlesDir := filepath.Join(GearBuildPath, m.RegistryIp+":"+m.RegistryPort+"/"+image.Repository+"-gear"+":"+image.Tag, "files")
    pusher, err := push.InitPusher(gFIlesDir, m.ManagerIp, m.ManagerPort, m.NoCleanUp)
    if err != nil {
        logrus.Fatal("Fail to init a pusher to push gear image...")
    }
    pusher.Push()

    cName := "docker"
    cArgs := []string{"push", m.RegistryIp+":"+m.RegistryPort+"/"+image.Repository+"-gear"+":"+image.Tag}
    cCmd := exec.Command(cName, cArgs...)
    if err := cCmd.Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Println("done!")

	return nil
}

func parseRegistry(registry string) (ip string, port string) {
	ipAndPort := strings.Split(registry, ":")

	if len(ipAndPort) != 2 {
		logger.Fatal("Invalid registry ip and port...")
	}

	ip = ipAndPort[0]
	port = ipAndPort[1]

	return
}

func walkDirs(dirs []string) ([]string, error) {
	var pathsToBeNoticed = []string{}
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
			// fail to get file info
			if f == nil {
				return err
			}

			if f.Mode().IsDir() {
				pathsToBeNoticed = append(pathsToBeNoticed, path)
			}

			return nil
		})
		if err != nil {
			logger.Warn("Fail to walk layers of image...")
			return nil, err
		}
	}

	return pathsToBeNoticed, nil
}