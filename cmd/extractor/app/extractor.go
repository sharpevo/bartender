package app

import (
	"automation/cmd/extractor/options"
	commonOptions "automation/cmd/options"
	"automation/internal/fsop"
	"automation/pkg/sshtrans"
	"automation/pkg/watchrecur"
	"excel"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
)

type ExtractCommand struct {
	LogOptions    *commonOptions.LogOptions
	ServerOptions *commonOptions.ServerOptions
	WatchOptions  *commonOptions.WatchOptions
	Options       *options.Options
	Recursive     bool
	Columns       []int
	Regexp        *regexp.Regexp
}

func NewExtractCommand() *ExtractCommand {
	return &ExtractCommand{
		LogOptions:    commonOptions.AttachLogOptions(flag.CommandLine),
		ServerOptions: commonOptions.AttachServerOptions(flag.CommandLine),
		WatchOptions:  commonOptions.AttachWatchOptions(flag.CommandLine),
		Options:       options.AttachOptions(flag.CommandLine),
	}
}

func (c *ExtractCommand) Validate() (err error) {
	flag.Parse()
	c.Columns, err = fsop.ConvertColumnIndices(c.Options.ColumnIndices)
	if err != nil {
		return err
	}
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
		}).Error("PRS")
	}
	return nil
}

func (c *ExtractCommand) Execute() error {
	if !c.Recursive {
		if err := c.process(c.Options.InputPath); err != nil {
			logrus.WithFields(logrus.Fields{
				"file":    c.Options.InputPath,
				"message": err.Error(),
			}).Error("PRS")
			return err
		}
		return nil
	}

	if !c.WatchOptions.Enabled {
		walkfunc := func(inputPath string, f os.FileInfo, err error) error {
			if !f.Mode().IsRegular() {
				return nil
			}
			if err := c.process(inputPath); err != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("PRS")
			}
			return nil
		}
		if err := filepath.Walk(c.Options.InputPath, walkfunc); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    c.Options.InputPath,
				"message": err.Error(),
			}).Error("PRS")
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
				}).Error("PRS")
			}
			return nil
		},
	)
	return nil
}

func (c *ExtractCommand) process(inputPath string) error {
	if !c.Regexp.MatchString(inputPath) {
		logrus.WithFields(logrus.Fields{
			"file": inputPath,
			"message": fmt.Sprintf(
				"file '%s' is not matched with pattern '%s'",
				inputPath,
				c.WatchOptions.FileNamePattern,
			),
		}).Warn("PRS")
		return nil
	}
	outputFile, err := c.extract(inputPath)
	if err != nil {
		return err
	}
	if !c.ServerOptions.Enabled {
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"file": inputPath,
	}).Info("NEW")
	remoteDir, remoteFileName := fsop.CustomRemoteFileNameAndDir(
		inputPath,
		c.ServerOptions.Directory,
		c.Options.OutputType,
	)
	if sshtrans.TransViaPassword(
		c.ServerOptions.HostKey,
		c.ServerOptions.UserName,
		c.ServerOptions.Password,
		outputFile,
		remoteFileName,
		remoteDir,
	) != nil {
		return err
	}
	return nil
}

func (c *ExtractCommand) extract(inputPath string) (outputFile string, err error) {
	data, err := excel.ExtractColumns(
		inputPath,
		c.Options.SheetIndex,
		c.Options.RowStartsAt,
		c.Options.RowEndsAt,
		c.Columns,
	)
	if err != nil {
		return outputFile, err
	}
	outputFile = fsop.MakeOutputFilePath(
		c.Options.OutputPath,
		inputPath,
		c.Options.OutputType,
	)
	switch c.Options.OutputType {
	case excel.OUTPUT_TYPE_CSV, excel.OUTPUT_TYPE_TXT:
		if excel.MakeFileCSV(
			outputFile,
			data,
			excel.SEPARATOR_TAB,
		) != nil {
			return outputFile, err
		}
	case excel.OUTPUT_TYPE_XLSX:
		if excel.MakeFileXLSX(
			outputFile,
			data,
			"sheet-0",
		) != nil {
			return outputFile, err
		}
	default:
		return outputFile, fmt.Errorf(
			"invalid file type '%v'",
			c.Options.OutputType,
		)
	}
	logrus.WithFields(logrus.Fields{
		"file": outputFile,
	}).Info("PRS")
	return outputFile, nil
}
