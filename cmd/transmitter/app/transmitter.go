package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	commonOptions "github.com/sharpevo/bartender/cmd/options"
	"github.com/sharpevo/bartender/cmd/transmitter/options"
	"github.com/sharpevo/bartender/internal/fsop"
	"github.com/sharpevo/bartender/pkg/messenger"
	"github.com/sharpevo/bartender/pkg/sshtrans"
	"github.com/sharpevo/bartender/pkg/watchrecur"

	"github.com/sirupsen/logrus"
)

type TransmitterCommand struct {
	LogOptions    *commonOptions.LogOptions
	ServerOptions *commonOptions.ServerOptions
	WatchOptions  *commonOptions.WatchOptions
	DingOptions   *commonOptions.DingOptions
	Options       *options.Options
	Regexp        *regexp.Regexp
	Recursive     bool
}

func NewTransmitterCommand() *TransmitterCommand {
	return &TransmitterCommand{
		LogOptions:    commonOptions.AttachLogOptions(flag.CommandLine),
		WatchOptions:  commonOptions.AttachWatchOptions(flag.CommandLine),
		ServerOptions: commonOptions.AttachServerOptions(flag.CommandLine),
		DingOptions:   commonOptions.AttachDingOptions(flag.CommandLine),
		Options:       options.AttachOptions(flag.CommandLine),
	}
}

func (c *TransmitterCommand) Validate() (err error) {
	flag.Parse()
	c.Recursive, err = fsop.IsDir(c.Options.InputPath)
	if err != nil {
		return err
	}
	switch c.LogOptions.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.WithFields(logrus.Fields{
		"logOptions":    commonOptions.Debug(c.LogOptions),
		"watchOptions":  commonOptions.Debug(c.WatchOptions),
		"serverOptions": commonOptions.Debug(c.ServerOptions),
		"options":       commonOptions.Debug(c.Options),
	}).Debug("LOG")
	c.Regexp, err = regexp.Compile(c.WatchOptions.FileNamePattern)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    c.Options.InputPath,
			"message": err.Error(),
		}).Error("TRS")
	}
	return nil
}

func (c *TransmitterCommand) Execute() error {
	if !c.Recursive {
		if err := c.process(c.Options.InputPath); err != nil {
			logrus.WithFields(logrus.Fields{
				"file":    c.Options.InputPath,
				"message": err.Error(),
			}).Error("TRS")
			messenger.Send(
				c.DingOptions.Token,
				fmt.Sprintf(
					"**Failed to send file after 10 retries**\n\n%s\n\n######%s",
					c.Options.InputPath,
					c.DingOptions.Source))
			return err
		}
		return nil
	}
	if !c.WatchOptions.Enabled {
		walkfunc := func(inputPath string, f os.FileInfo, err error) error {
			if !f.Mode().IsRegular() {
				return nil
			}
			if c.process(inputPath) != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("TRS")
			}
			return nil
		}
		if err := filepath.Walk(c.Options.InputPath, walkfunc); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    c.Options.InputPath,
				"message": err.Error(),
			}).Error("TRS")
		}
		return nil
	}

	watchrecur.Watch(
		c.Options.InputPath,
		c.WatchOptions.Interval,
		func(inputPath string) error {
			if err := c.process(inputPath); err != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("TRS")
			}
			return nil
		},
		func(inputPath string) error {
			logrus.WithFields(logrus.Fields{
				"file":    inputPath,
				"message": "delete event is omitted",
			}).Warn("TRS")
			return nil
		},
	)
	return nil
}

func (c *TransmitterCommand) process(inputPath string) error {
	if !c.Regexp.MatchString(inputPath) {
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf(
				"file '%s' is not matched with pattern '%s'",
				inputPath,
				c.WatchOptions.FileNamePattern,
			),
		}).Warn("TRS")
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"file": inputPath,
	}).Info("TRS")
	if !c.ServerOptions.Enabled {
		return nil
	}
	remoteRel, err := fsop.GetRelativePath(c.WatchOptions.InputPath, inputPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file":    inputPath,
			"message": err.Error(),
		}).Error("TRS")
		return err
	}
	remoteDir := filepath.Join(c.ServerOptions.Directory, remoteRel)
	_, fileName := filepath.Split(inputPath)
	return sshtrans.TransViaPassword(
		c.ServerOptions.HostKey,
		c.ServerOptions.UserName,
		c.ServerOptions.Password,
		inputPath,
		fileName,
		remoteDir,
	)
}
