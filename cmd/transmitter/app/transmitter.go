package app

import (
	commonOptions "automation/cmd/options"
	"automation/cmd/transmitter/options"
	"automation/pkg/sshtrans"
	"automation/pkg/watchrecur"
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
)

type TransmitterCommand struct {
	LogOptions      *commonOptions.LogOptions
	ServerOptions   *commonOptions.ServerOptions
	WatchOptions    *commonOptions.WatchOptions
	TransferOptions *options.TransferOptions
	Regexp          *regexp.Regexp
}

func NewTransmitterCommand() *TransmitterCommand {
	return &TransmitterCommand{
		LogOptions:      commonOptions.AttachLogOptions(flag.CommandLine),
		WatchOptions:    commonOptions.AttachWatchOptions(flag.CommandLine),
		ServerOptions:   commonOptions.AttachServerOptions(flag.CommandLine),
		TransferOptions: options.AttachTransferOptions(flag.CommandLine),
	}
}

func (c *TransmitterCommand) Validate() (err error) {
	flag.Parse()
	switch c.LogOptions.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.WithFields(logrus.Fields{
		"logOptions":      commonOptions.Debug(c.LogOptions),
		"watchOptions":    commonOptions.Debug(c.WatchOptions),
		"serverOptions":   commonOptions.Debug(c.ServerOptions),
		"transferOptions": commonOptions.Debug(c.TransferOptions),
	}).Debug("LOG")
	c.Regexp, err = regexp.Compile(c.WatchOptions.FileNamePattern)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    c.TransferOptions.InputPath,
			"message": err.Error(),
		}).Error("TRS")
	}
	return nil
}

func (c *TransmitterCommand) Execute() error {
	if !c.WatchOptions.Enabled {
		walkfunc := func(inputPath string, f os.FileInfo, err error) error {
			if !f.Mode().IsRegular() || !c.Regexp.MatchString(inputPath) {
				return nil
			}
			logrus.WithFields(logrus.Fields{
				"file": inputPath,
			}).Info("TRS")
			if c.HandleTransfer(inputPath) != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("TRS")
			}
			return nil
		}
		if err := filepath.Walk(c.TransferOptions.InputPath, walkfunc); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    c.TransferOptions.InputPath,
				"message": err.Error(),
			}).Error("TRS")
		}
		return nil
	}

	watchrecur.Watch(
		c.TransferOptions.InputPath,
		c.WatchOptions.Interval,
		func(inputPath string) error {
			if !c.Regexp.MatchString(inputPath) {
				return nil
			}
			if err := c.HandleTransfer(inputPath); err != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("TRS")
			}
			return nil
		},
	)
	return nil
}

func (c *TransmitterCommand) HandleTransfer(inputPath string) error {
	if !c.ServerOptions.Enabled {
		return nil
	}
	_, fileName := filepath.Split(inputPath)
	return sshtrans.TransViaPassword(
		c.ServerOptions.HostKey,
		c.ServerOptions.UserName,
		c.ServerOptions.Password,
		inputPath,
		fileName,
		c.ServerOptions.Directory,
	)
}
