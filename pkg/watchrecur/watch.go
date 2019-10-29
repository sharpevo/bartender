package watchrecur

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Callback func(inputPath string) error

func (c Callback) call(inputPath string) error {
	return c(inputPath)
}

var (
	watcher    *fsnotify.Watcher
	terminated = make(chan struct{})
	fileLock   = sync.RWMutex{}
	benchMap   = make(map[string]time.Time)
	benchLock  = sync.RWMutex{}
)

func Watch(
	inputPath string,
	interval time.Duration,
	callback Callback,
) {
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()
	watch(inputPath, false)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// TODO: event overflow
				if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
					fileLock.Lock()
					filePath := event.Name
					go func(inputPath string) {
						benchLock.Lock()
						tryRemoveFromBench(inputPath)
						benchLock.Unlock()
						logrus.WithFields(logrus.Fields{
							"message": fmt.Sprintf("DET '%s'", inputPath),
						}).Debug("WCH")
						fileLock.Unlock()
						if callback.call(inputPath) != nil {
							terminated <- struct{}{}
						}
					}(filePath)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					filePath := event.Name
					if !isDir(filePath) {
						continue
					}
					fileLock.Lock()
					watch(filePath, true)
					fileLock.Unlock()
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
	go scanBench(callback, interval)
	<-terminated
}

func watch(inputPath string, benchable bool) error {
	logrus.WithFields(logrus.Fields{
		"message": fmt.Sprintf("watch '%s'", inputPath),
	}).Debug("WCH")
	watcher.Add(inputPath)
	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		filePath := filepath.Join(inputPath, f.Name())
		if f.Mode().IsDir() {
			watch(filePath, benchable)
			continue
		}
		if benchable && f.Mode().IsRegular() {
			benchLock.Lock()
			benchMap[filePath] = f.ModTime()
			benchLock.Unlock()
			logrus.WithFields(logrus.Fields{
				"message": fmt.Sprintf("bench '%s'", filePath),
			}).Debug("WCH")
			continue
		}
	}
	return nil
}

func tryRemoveFromBench(inputPath string) {
	if _, ok := benchMap[inputPath]; ok {
		delete(benchMap, inputPath)
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf("unbench '%s'", inputPath),
		}).Debug("WCH")
	}
}

func scanBench(callback Callback, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		benchLock.Lock()
		earlist := time.Now().Add(-interval)
		for filePath, timestamp := range benchMap {
			if timestamp.Before(earlist) {
				f, _ := os.Stat(filePath)
				tryRemoveFromBench(filePath)
				if !f.ModTime().Equal(timestamp) {
					logrus.WithFields(logrus.Fields{
						"message": fmt.Sprintf("IGN '%s': %v",
							filePath,
							timestamp.Format("2006-01-02_15:04:05"),
						),
					}).Debug("WCH")
					continue
				}
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("EXP '%s': %v",
						filePath,
						timestamp.Format("2006-01-02_15:04:05"),
					),
				}).Debug("WCH")
				if err := callback.call(filePath); err != nil {
					terminated <- struct{}{}
					ticker.Stop()
				}
			}
		}
		benchLock.Unlock()
	}
}

func isDir(inputPath string) bool {
	f, _ := os.Stat(inputPath)
	return f.Mode().IsDir()
}
