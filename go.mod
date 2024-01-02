module github.com/sharpevo/bartender

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/pkg/sftp v1.11.0
	github.com/radovskyb/watcher v1.0.7
	github.com/sharpevo/xlsxutil v0.0.0-20200818064812-83017a0cbbe7
	github.com/sirupsen/logrus v1.6.0
	github.com/tealeg/xlsx/v3 v3.2.0 // indirect
	golang.org/x/crypto v0.17.0
)

replace (
	github.com/fsnotify/fsnotify => /home/yang/go/src/github.com/fsnotify/fsnotify
	github.com/sharpevo/testutil => /home/yang/go/src/github.com/sharpevo/testutil
	github.com/sharpevo/xlsxutil => /home/yang/go/src/github.com/sharpevo/xlsxutil
)
