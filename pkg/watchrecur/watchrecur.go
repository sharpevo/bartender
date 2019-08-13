package watchrecur

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

var (
	watcher    *fsnotify.Watcher
	terminated = make(chan struct{})
)

type Callback func(inputPath string) error

func NewWatcher() *fsnotify.Watcher {
	watcher, _ = fsnotify.NewWatcher()
	return watcher
}

func Watch(
	inputPath string,
	interval time.Duration,
	callback Callback,
) {
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
					if callback(event.Name) != nil {
						terminated <- struct{}{}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithFields(logrus.Fields{
					"message": err.Error(),
				}).Error("WCH")
			}
		}
	}()
	go scan(inputPath, interval)
	<-terminated
}

func scan(inputPath string, duration time.Duration) {
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		<-done
		if err := filepath.Walk(inputPath, addToWatcher); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    inputPath,
				"message": err.Error(),
			}).Error("WCH")
		}
		go func() {
			done <- struct{}{}
		}()
	}
}

func addToWatcher(inputPath string, f os.FileInfo, err error) error {
	if f.Mode().IsDir() {
		logrus.WithFields(logrus.Fields{
			"file": inputPath,
		}).Debug("WCH")
		return watcher.Add(inputPath)
	}
	return nil
}
