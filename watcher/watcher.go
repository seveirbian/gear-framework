package watcher

import (
	"os"
	"fmt"
	// "sync"

	"net/http"
	"net/url"

	"github.com/jandre/fanotify"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.WithField("gear", "watcher")
	procFsFdInfo   = "/proc/self/fd/%d"
)

type Event struct {
	Path string
	Op string
}

type Server struct {
	Ip string
	Port string
}

func Watch(mountPoint string, server Server) {
	watcher, err := fanotify.Initialize(fanotify.FAN_CLASS_NOTIF, os.O_RDONLY)
	if err != nil {
		logger.Warnf("Fail to Initialize watcher for %v", err)
	}

	err = watcher.Mark(fanotify.FAN_MARK_ADD|fanotify.FAN_MARK_MOUNT,
		fanotify.FAN_MODIFY|fanotify.FAN_ACCESS|fanotify.FAN_OPEN, -1, mountPoint)

	if err != nil {
		logger.Warnf("Fail to mark for %v", err)
	}

	func() {
		for {
			data, err := watcher.GetEvent()
			if err != nil {
				logger.Warnf("Fail to get event for %v", err)
			}

			if (data.Mask & fanotify.FAN_Q_OVERFLOW) == fanotify.FAN_Q_OVERFLOW {
				logger.Debugf("fanmon: collector - overflow event")
				continue
			}

			path, err := os.Readlink(fmt.Sprintf(procFsFdInfo, data.File.Fd()))
			if err != nil {
				logger.Warnf("Fail to readlink for %v", err)
			}
			data.File.Close()

			fmt.Println(path)

			// 向服务器返回获取到的信息
			resp, err := http.PostForm(server.Ip + ":" + server.Port + "/event", url.Values{"path":{path}})
			if err != nil {
				logger.Warnf("Fail to post server for %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				logger.Warnf("Status Code is not ok...")
			}
		}
	}()
}