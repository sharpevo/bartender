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
	"log"
	"os"
	"path/filepath"
)

type ExtractCommand struct {
	LogOptions    *commonOptions.LogOptions
	ServerOptions *commonOptions.ServerOptions
	WatchOptions  *commonOptions.WatchOptions
	ParseOptions  *options.ParseOptions
	Recursive     bool
	Columns       []int
}

func NewExtractCommand() *ExtractCommand {
	return &ExtractCommand{
		LogOptions:    commonOptions.AttachLogOptions(flag.CommandLine),
		ServerOptions: commonOptions.AttachServerOptions(flag.CommandLine),
		WatchOptions:  commonOptions.AttachWatchOptions(flag.CommandLine),
		ParseOptions:  options.AttachParseOptions(flag.CommandLine),
	}
}

func (c *ExtractCommand) Validate() (err error) {
	flag.Parse()
	c.Columns, err = fsop.ConvertColumnIndices(c.ParseOptions.ColumnIndices)
	if err != nil {
		return err
	}
	c.Recursive, err = fsop.IsDir(c.ParseOptions.InputPath)
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
		"parseOptions":  commonOptions.Debug(c.ParseOptions),
	}).Debug("LOG")
	return nil
}

func (c *ExtractCommand) Execute() error {
	if !c.Recursive {
		outputFile, err := c.Extract(c.ParseOptions.InputPath)
		if err != nil {
			return err
		}
		log.Println(outputFile)
		return nil
	}

	if !c.WatchOptions.Enabled {
		walkfunc := func(inputPath string, f os.FileInfo, err error) error {
			if !f.Mode().IsRegular() {
				return nil
			}
			logrus.WithFields(logrus.Fields{
				"file": inputPath,
			}).Info("NEW")
			if err := c.HandleParse(inputPath); err != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("PRS")
			}
			return nil
		}
		if err := filepath.Walk(c.ParseOptions.InputPath, walkfunc); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    c.ParseOptions.InputPath,
				"message": err.Error(),
			}).Error("PRS")
		}
		return nil
	}

	watchrecur.Watch(
		c.ParseOptions.InputPath,
		c.WatchOptions.Interval,
		func(inputPath string) error {
			logrus.WithFields(logrus.Fields{
				"file": inputPath,
			}).Info("NEW")
			if err := c.HandleParse(inputPath); err != nil {
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

func (c *ExtractCommand) HandleParse(
	inputPath string,
) error {
	outputFile, err := c.Extract(inputPath)
	if err != nil {
		return err
	}
	if !c.ServerOptions.Enabled {
		return nil
	}
	remoteDir, remoteFileName := fsop.CustomRemoteFileNameAndDir(
		inputPath,
		c.ServerOptions.Directory,
		c.ParseOptions.OutputType,
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

func (c *ExtractCommand) Extract(inputPath string) (outputFile string, err error) {
	data, err := excel.ExtractColumns(
		inputPath,
		c.ParseOptions.SheetIndex,
		c.ParseOptions.RowStartsAt,
		c.ParseOptions.RowEndsAt,
		c.Columns,
	)
	if err != nil {
		return outputFile, err
	}
	outputFile = fsop.MakeOutputFilePath(
		c.ParseOptions.OutputPath,
		inputPath,
		c.ParseOptions.OutputType,
	)
	switch c.ParseOptions.OutputType {
	case "csv", "txt":
		if excel.MakeFileCSV(
			outputFile,
			data,
			excel.SEPARATOR_TAB,
		) != nil {
			return outputFile, err
		}
	case "xlsx":
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
			c.ParseOptions.OutputType,
		)
	}
	logrus.WithFields(logrus.Fields{
		"file": outputFile,
	}).Info("PRS")
	return outputFile, nil
}
