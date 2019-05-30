package main

import (
	"os"
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
  	"github.com/sirupsen/logrus"
)

var (
  logger = logrus.WithField("gear", "fs")
)

func main () {
  stop := make(chan bool)
  out := make(chan fsnotify.Event, 50)
  
  dirs, _ := walkDirs([]string{"/var/lib/docker/overlay2/719a6fa9d5fdd76304b1ef5d76d4f129b2b82c677f497e7b3db2f1c97ab48cbe-init/diff", "/var/lib/docker/overlay2/1edec9b69909bb4057c438e99f9cc93b6706434a718cf92959224e062ece7a7c/diff"})
  
  Watch(dirs, stop, out)
  
  for event := range out {
    fmt.Println(event)
  }
}

func Watch(paths []string, stop chan bool, out chan fsnotify.Event) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Warnf("Fail to create watcher for %v", err)
	}

	go func() {
		for {
			select {
			case <- stop:
				return
			case event, ok := <- watcher.Events:
				fmt.Println(event)
				if !ok {
					logger.Warnf("Fail to catch event...")
				}
				fmt.Println(event)
				out <- event
			case err, ok := <- watcher.Errors:
				if !ok {
					return
				}
				logger.Warnf("Watcher error: ", err)
			}
		}
	}()

	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			logger.Warnf("Fail to add path %s for %v", path, err)
		}
	}
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