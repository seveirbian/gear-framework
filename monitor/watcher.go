package monitor

import (
	// "os"
	"fmt"
	// "sync"

	"github.com/fsnotify/fsnotify"
)

var (
	
)

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