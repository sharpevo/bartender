package watchrecur

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	watcher      *fsnotify.Watcher
	terminated   = make(chan struct{})
	benchMap     = make(map[string]time.Time)
	benchLock    = sync.RWMutex{}
	callbackLock = sync.RWMutex{}
)

type Callback func(inputPath string) error

func (c Callback) call(inputPath string) error {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	return c(inputPath)
}

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
					filePath := event.Name
					benchLock.Lock()
					tryRemoveFromBench(filePath)
					benchLock.Unlock()
					logrus.WithFields(logrus.Fields{
						"message": fmt.Sprintf("DET '%s'", filePath),
					}).Debug("WCH")
					if callback.call(filePath) != nil {
						terminated <- struct{}{}
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					filePath := event.Name
					f, _ := os.Stat(filePath)
					if f.Mode().IsDir() {
						var err error
						addToWatcher(filePath, f, err)
						benchLock.Lock()
						scan(filePath)
						benchLock.Unlock()
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
	go scanDirectories(inputPath)
	go scanBench(callback, interval)
	<-terminated
}

func scan(inputPath string) {
	if err := filepath.Walk(inputPath, add([]string{inputPath})); err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    inputPath,
			"message": err.Error(),
		}).Error("WCH")
	}
}

func add(ignores []string) filepath.WalkFunc {
	return func(inputPath string, f os.FileInfo, err error) error {
		for _, ignore := range ignores {
			if inputPath == ignore {
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("ignore directory '%s'", inputPath),
				}).Debug("WCH")
				return nil
			}
		}
		if f.Mode().IsDir() {
			addToWatcher(inputPath, f, err)
			scan(inputPath)
			return nil
		}
		if f.Mode().IsRegular() {
			//benchLock.Lock()
			benchMap[inputPath] = time.Now()
			logrus.WithFields(logrus.Fields{
				"message": fmt.Sprintf("bench '%s'", inputPath),
			}).Debug("WCH")
			//benchLock.Unlock()
			return nil
		}
		return nil
	}
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
				tryRemoveFromBench(filePath)
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

func scanDirectories(inputPath string) {
	if err := filepath.Walk(inputPath, addToWatcher); err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    inputPath,
			"message": err.Error(),
		}).Error("WCH")
	}
}

func addToWatcher(inputPath string, f os.FileInfo, err error) error {
	if f.Mode().IsDir() {
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf("watch '%s'", inputPath),
		}).Debug("WCH")
		return watcher.Add(inputPath)
	}
	return nil
}
