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
	watcher    *fsnotify.Watcher
	terminated = make(chan struct{})
	benchMap   = make(map[string]time.Time)
	benchLock  = sync.RWMutex{}
)

type Callback func(inputPath string) error

func NewWatcher() *fsnotify.Watcher {
	watcher, _ = fsnotify.NewWatcher()
	return watcher
}

func Watch(
	inputPath string,
	expiration time.Duration,
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
					if callback(filePath) != nil {
						terminated <- struct{}{}
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					filePath := event.Name
					f, _ := os.Stat(filePath)
					if f.Mode().IsDir() {
						var err error
						addToWatcher(filePath, f, err)
						scan(filePath)
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
	go scanBench(callback, expiration)
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
			benchLock.Lock()
			benchMap[inputPath] = time.Now()
			logrus.WithFields(logrus.Fields{
				"message": fmt.Sprintf("add '%s' to bench", inputPath),
			}).Debug("WCH")
			benchLock.Unlock()
			return nil
		}
		return nil
	}
}

func tryRemoveFromBench(inputPath string) {
	if _, ok := benchMap[inputPath]; ok {
		delete(benchMap, inputPath)
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf("remove '%s' from bench", inputPath),
		}).Debug("WCH")
	}
}

func scanBench(callback Callback, expiration time.Duration) {
	ticker := time.NewTicker(expiration)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		benchLock.Lock()
		earlist := time.Now().Add(-expiration)
		for filePath, timestamp := range benchMap {
			if timestamp.Before(earlist) {
				tryRemoveFromBench(filePath)
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("EXP '%s': %v",
						filePath,
						timestamp.Format("2006-01-02_15:04:05"),
					),
				}).Debug("WCH")
				if err := callback(filePath); err != nil {
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

func xscan(inputPath string, duration time.Duration) {
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
			"file":    inputPath,
			"message": fmt.Sprintf("add '%v' to watcher", inputPath),
		}).Debug("WCH")
		return watcher.Add(inputPath)
	}
	return nil
}
