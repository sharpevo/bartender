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
	callbackMap  = make(map[string]map[string]bool)
	callbackLock = sync.RWMutex{}
)

type Callback func(inputPath string) error

func initCallbackMap(inputPath string) {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	cm := make(map[string]bool)
	callbackMap[callbackMapKey(inputPath)] = cm
	logrus.WithFields(logrus.Fields{
		"message": fmt.Sprintf("callbackmap init '%s'", inputPath),
	}).Debug("WCH")
}

func clearCallbackMap(inputPath string) {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	delete(callbackMap, callbackMapKey(inputPath))
	logrus.WithFields(logrus.Fields{
		"message": fmt.Sprintf("callbackmap destroy '%s'", inputPath),
	}).Debug("WCH")
}

func callbackMapKey(inputPath string) string {
	return filepath.Join(inputPath, "*")
}

func addCallbackMap(inputPath string) {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	if len(callbackMap) > 0 {
		for pattern, cm := range callbackMap {
			if matched, _ := filepath.Match(pattern, inputPath); matched {
				cm[inputPath] = true
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("callbackmap add '%s'", inputPath),
				}).Debug("WCH")
			}
		}
	}
}

func inCallbackMap(inputPath string) bool {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	for _, cm := range callbackMap {
		if cm[inputPath] {
			return true
		}
	}
	return false
}

func (c Callback) call(inputPath string) error {
	addCallbackMap(inputPath)
	return c(inputPath)
}

func Watch(
	inputPath string,
	interval time.Duration,
	callback Callback,
) {
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()
	watchDirectory(inputPath)
	go scanDirectory(inputPath, false)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// TODO: event overflow
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
					if !isDir(filePath) {
						continue
					}
					logrus.WithFields(logrus.Fields{
						"message": fmt.Sprintf("DET create '%s'", filePath),
					}).Debug("WCH")
					watchDirectory(filePath)
					scanDirectory(filePath, true)
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

func scanDirectory(inputPath string, benchable bool) {
	logrus.WithFields(logrus.Fields{
		"message": fmt.Sprintf("scan '%s'", inputPath),
	}).Debug("WCH")
	initCallbackMap(inputPath)
	if !isDir(inputPath) {
		return
	}
	if err := filepath.Walk(inputPath, addWatch([]string{inputPath}, benchable)); err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    inputPath,
			"message": err.Error(),
		}).Error("WCH")
	}
	clearCallbackMap(inputPath)
}

func addWatch(ignores []string, benchable bool) filepath.WalkFunc {
	return func(inputPath string, f os.FileInfo, err error) error {
		for _, ignore := range ignores {
			if inputPath == ignore {
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("ignore '%s'", inputPath),
				}).Debug("WCH")
				return nil
			}
		}
		if f.Mode().IsDir() {
			watchDirectory(inputPath)
			scanDirectory(inputPath, benchable)
			return filepath.SkipDir
		}
		if f.Mode().IsRegular() && benchable {
			if inCallbackMap(inputPath) {
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("not bench '%s'", inputPath),
				}).Debug("WCH")
				return nil
			}
			benchLock.Lock()
			benchMap[inputPath] = time.Now()
			benchLock.Unlock()
			logrus.WithFields(logrus.Fields{
				"message": fmt.Sprintf("bench '%s'", inputPath),
			}).Debug("WCH")
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

func watchDirectory(inputPath string) error {
	if !isDir(inputPath) {
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"message": fmt.Sprintf("watch '%s'", inputPath),
	}).Debug("WCH")
	// TODO: ignore EINVAL
	return watcher.Add(inputPath)
}

func isDir(inputPath string) bool {
	f, _ := os.Stat(inputPath)
	return f.Mode().IsDir()
}
