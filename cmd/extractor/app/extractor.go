package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sharpevo/bartender/cmd/extractor/options"
	commonOptions "github.com/sharpevo/bartender/cmd/options"
	"github.com/sharpevo/bartender/internal/fsop"
	"github.com/sharpevo/bartender/pkg/messenger"
	"github.com/sharpevo/bartender/pkg/sshtrans"
	"github.com/sharpevo/bartender/pkg/watchrecur"

	"github.com/sharpevo/xlsxutil"
	"github.com/sirupsen/logrus"
)

type ExtractCommand struct {
	LogOptions    *commonOptions.LogOptions
	ServerOptions *commonOptions.ServerOptions
	WatchOptions  *commonOptions.WatchOptions
	DingOptions   *commonOptions.DingOptions
	Options       *options.Options
	Recursive     bool
	Columns       []int
	Regexp        *regexp.Regexp
	RegexpExtract *regexp.Regexp
}

func NewExtractCommand() *ExtractCommand {
	return &ExtractCommand{
		LogOptions:    commonOptions.AttachLogOptions(flag.CommandLine),
		ServerOptions: commonOptions.AttachServerOptions(flag.CommandLine),
		WatchOptions:  commonOptions.AttachWatchOptions(flag.CommandLine),
		DingOptions:   commonOptions.AttachDingOptions(flag.CommandLine),
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
	c.RegexpExtract, err = regexp.Compile(c.Options.ExtractPattern)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path":    c.Options.ExtractPattern,
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
		func(inputPath string) error {
			if err := c.remove(inputPath); err != nil {
				logrus.WithFields(logrus.Fields{
					"file":    inputPath,
					"message": err.Error(),
				}).Error("RMV")
			}
			return nil
		},
	)
	return nil
}

func (c *ExtractCommand) process(inputPath string) error {
	if !c.Regexp.MatchString(inputPath) {
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf(
				"file '%s' is not matched with pattern '%s'",
				inputPath,
				c.WatchOptions.FileNamePattern,
			),
		}).Warn("PRS")
		return nil
	}
	var outputFile string
	var err error
	if c.RegexpExtract.MatchString(inputPath) {
		outputFile, err = c.extract(inputPath)
		if err != nil {
			messenger.Send(
				c.DingOptions.Token,
				fmt.Sprintf(
					"**Failed to extract xlsx file**\n\n%s\n\n###### %s",
					inputPath, c.DingOptions.Source))
			return err
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"message": fmt.Sprintf(
				"file '%s' will transfer directly",
				inputPath,
			),
		}).Info("PRS")
		outputFile = inputPath
	}
	if !c.ServerOptions.Enabled {
		return nil
	}
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

func (c *ExtractCommand) remove(inputPath string) error {
	remoteDir, _ := fsop.CustomRemoteFileNameAndDir(
		inputPath,
		c.ServerOptions.Directory,
		c.Options.OutputType,
	)
	return sshtrans.RemoveViaPassword(
		c.ServerOptions.HostKey,
		c.ServerOptions.UserName,
		c.ServerOptions.Password,
		remoteDir,
	)
}

func (c *ExtractCommand) extract(inputPath string) (outputFile string, err error) {
	index := c.Options.SheetIndex
	if filepath.Ext(inputPath) == ".xlsx" {
		logrus.Infof("change the sheet index to 0: %s", filepath.Base(inputPath))
		index = 0
	}
	data, err := xlsxutil.ExtractColumns(
		inputPath,
		index,
		c.Options.RowStartsAt,
		c.Options.RowEndsAt,
		c.Columns,
	)
	if err != nil {
		return outputFile, err
	}
	data = trimSpacesAndTabs(data)
	outputFile = fsop.MakeOutputFilePath(
		c.Options.OutputPath,
		inputPath,
		c.Options.OutputType,
	)
	dir, _ := filepath.Split(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return outputFile, err
	}
	switch c.Options.OutputType {
	case xlsxutil.OUTPUT_TYPE_CSV, xlsxutil.OUTPUT_TYPE_TXT:
		if xlsxutil.MakeFileCSV(
			outputFile,
			data,
			xlsxutil.SEPARATOR_TAB,
		) != nil {
			return outputFile, err
		}
	case xlsxutil.OUTPUT_TYPE_XLSX:
		if xlsxutil.MakeFileXLSX(
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
		"message": fmt.Sprintf(
			"extract '%s' to '%s'",
			inputPath,
			outputFile,
		),
	}).Info("PRS")
	return outputFile, nil
}

func trimSpacesAndTabs(rows [][]string) [][]string {
	output := [][]string{}
	for _, row := range rows {
		cells := []string{}
		for _, cell := range row {
			cells = append(cells,
				strings.Replace(strings.Replace(cell, " ", "", -1), "\t", "", -1))
		}
		output = append(output, cells)
	}
	return output
}
