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

func NewExtractCommand() {
	logOptions := commonOptions.AttachLogOptions(flag.CommandLine)
	serverOptions := commonOptions.AttachServerOptions(flag.CommandLine)
	watchOptions := commonOptions.AttachWatchOptions(flag.CommandLine)
	parseOptions := options.AttachParseOptions(flag.CommandLine)
	flag.Parse()
	if flag.Parsed() {
		columns, err := fsop.ConvertColumnIndices(parseOptions.ColumnIndices)
		if err != nil {
			fmt.Println(err)
			return
		}
		isFolder, err := fsop.IsDir(parseOptions.InputPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		switch logOptions.Level {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}

		logrus.WithFields(logrus.Fields{
			"logOptions":    commonOptions.Debug(logOptions),
			"watchOptions":  commonOptions.Debug(watchOptions),
			"serverOptions": commonOptions.Debug(serverOptions),
			"parseOptions":  commonOptions.Debug(parseOptions),
		}).Debug("LOG")
		if !isFolder {
			if outputFile, err := Extract(
				parseOptions.InputPath,
				parseOptions.SheetIndex,
				parseOptions.RowStartsAt,
				parseOptions.RowEndsAt,
				columns,
				parseOptions.OutputPath,
				parseOptions.OutputType,
			); err != nil {
				log.Println(err)
				return
			} else {
				log.Println(outputFile)
				return
			}
			return
		}
		// is folder
		if !watchOptions.Enabled {
			if err := filepath.Walk(
				parseOptions.InputPath,
				func(
					inputPath string,
					f os.FileInfo,
					err error,
				) error {
					if !f.Mode().IsRegular() {
						return nil
					}
					logrus.WithFields(logrus.Fields{
						"file": inputPath,
					}).Info("NEW")
					if HandleParse(
						inputPath,
						columns,
						parseOptions,
						serverOptions,
					) != nil {
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": err.Error(),
						}).Error("PRS")
					}
					return nil
				}); err != nil {
				logrus.WithFields(logrus.Fields{
					"path":    parseOptions.InputPath,
					"message": err.Error(),
				}).Error("PRS")
			}
			return
		}
		watchrecur.Watch(
			parseOptions.InputPath,
			watchOptions.Interval,
			func(inputPath string) error {
				logrus.WithFields(logrus.Fields{
					"file": inputPath,
				}).Info("NEW")
				if HandleParse(
					inputPath,
					columns,
					parseOptions,
					serverOptions,
				) != nil {
					logrus.WithFields(logrus.Fields{
						"file":    inputPath,
						"message": err.Error(),
					}).Error("PRS")
				}
				return nil
			},
		)
	}
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
