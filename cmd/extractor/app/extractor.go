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
		outputFile, err := Extract(
			c.ParseOptions.InputPath,
			c.ParseOptions.SheetIndex,
			c.ParseOptions.RowStartsAt,
			c.ParseOptions.RowEndsAt,
			c.Columns,
			c.ParseOptions.OutputPath,
			c.ParseOptions.OutputType,
		)
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
			if HandleParse(
				inputPath,
				c.Columns,
				c.ParseOptions,
				c.ServerOptions,
			) != nil {
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
			if err := HandleParse(
				inputPath,
				c.Columns,
				c.ParseOptions,
				c.ServerOptions,
			); err != nil {
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

func HandleParse(
	inputPath string,
	columns []int,
	parseOptions *options.ParseOptions,
	serverOptions *commonOptions.ServerOptions,
) error {
	outputFile, err := Extract(
		inputPath,
		parseOptions.SheetIndex,
		parseOptions.RowStartsAt,
		parseOptions.RowEndsAt,
		columns,
		parseOptions.OutputPath,
		parseOptions.OutputType,
	)
	if err != nil {
		return err
	}
	if !serverOptions.Enabled {
		return nil
	}
	remoteDir, remoteFileName := fsop.CustomRemoteFileNameAndDir(
		inputPath,
		serverOptions.Directory,
		parseOptions.OutputType,
	)
	if sshtrans.TransViaPassword(
		serverOptions.HostKey,
		serverOptions.UserName,
		serverOptions.Password,
		outputFile,
		remoteFileName,
		remoteDir,
	) != nil {
		return err
	}
	return nil
}

func Extract(
	fileName string,
	sheetIndex int,
	rowStartsAt int,
	rowEndsAt int,
	columnIndices []int,
	outputPath string,
	outputType string,
) (outputFile string, err error) {
	data, err := excel.ExtractColumns(
		fileName,
		sheetIndex,
		rowStartsAt,
		rowEndsAt,
		columnIndices,
	)
	if err != nil {
		return outputFile, err
	}
	outputFile = fsop.MakeOutputFilePath(
		outputPath,
		fileName,
		outputType,
	)
	switch outputType {
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
			outputType,
		)
	}
	logrus.WithFields(logrus.Fields{
		"file": outputFile,
	}).Info("PRS")
	return outputFile, nil
}
