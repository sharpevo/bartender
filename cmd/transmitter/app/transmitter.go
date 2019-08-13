package app

import (
	commonOptions "automation/cmd/options"
	"automation/cmd/transmitter/options"
	"regexp"
	//"automation/parser/internal/fsop"
	"automation/pkg/sshtrans"
	"automation/pkg/watchrecur"
	"flag"
	//"fmt"
	"github.com/sirupsen/logrus"
	//"log"
	"os"
	"path/filepath"
)

func NewTransmitterCommand() {
	logOptions := commonOptions.AttachLogOptions(flag.CommandLine)
	watchOptions := commonOptions.AttachWatchOptions(flag.CommandLine)
	serverOptions := commonOptions.AttachServerOptions(flag.CommandLine)
	transferOptions := options.AttachTransferOptions(flag.CommandLine)
	flag.Parse()
	if flag.Parsed() {
		switch logOptions.Level {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
		logrus.WithFields(logrus.Fields{
			"logOptions":      commonOptions.Debug(logOptions),
			"watchOptions":    commonOptions.Debug(watchOptions),
			"serverOptions":   commonOptions.Debug(serverOptions),
			"transferOptions": commonOptions.Debug(transferOptions),
		}).Debug("LOG")
		r, err := regexp.Compile(watchOptions.FileNamePattern)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    transferOptions.InputPath,
				"message": err.Error(),
			}).Error("TRS")
		}
		if !watchOptions.Enabled {
			if err := filepath.Walk(
				transferOptions.InputPath,
				func(
					inputPath string,
					f os.FileInfo,
					err error,
				) error {
					if !f.Mode().IsRegular() ||
						!r.MatchString(inputPath) {
						return nil
					}
					logrus.WithFields(logrus.Fields{
						"file": inputPath,
					}).Info("TRS")
					if HandleTransfer(inputPath, serverOptions) != nil {
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": err.Error(),
						}).Error("TRS")
					} else {
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": "sent",
						}).Info("TRS")
					}
					return nil
				}); err != nil {
				logrus.WithFields(logrus.Fields{
					"path":    transferOptions.InputPath,
					"message": err.Error(),
				}).Error("TRS")
			}
			return
		}
		watchrecur.Watch(
			transferOptions.InputPath,
			watchOptions.Interval,
			func(inputPath string) error {
				if !r.MatchString(inputPath) {
					return nil
				}
				if HandleTransfer(inputPath, serverOptions) != nil {
					logrus.WithFields(logrus.Fields{
						"file":    inputPath,
						"message": err.Error(),
					}).Error("TRS")
				} else {
					logrus.WithFields(logrus.Fields{
						"file":    inputPath,
						"message": "sent",
					}).Info("TRS")
				}
				return nil
			},
		)
	}
}

func HandleTransfer(
	inputPath string,
	transferServerOptions *commonOptions.ServerOptions,
) error {
	if !transferServerOptions.Enabled {
		return nil
	}
	_, fileName := filepath.Split(inputPath)
	return sshtrans.TransViaPassword(
		transferServerOptions.HostKey,
		transferServerOptions.UserName,
		transferServerOptions.Password,
		inputPath,
		fileName,
		transferServerOptions.Directory,
	)
}
