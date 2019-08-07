//
// parse -file ../20190722-Nova314--分析需求表-贝瑞\(1\).xlsm -path /tmp/ -from 2 -sheet 1 -to -1 -type csv
// watch -directory /tmp/watcher -from 2 -sheet 1 -to -1 -type csv
// send -username wuy -password igenetech -hostkey "192.168.1.96 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBLjYGzkYWF+a1KV2NDjEtjzfa0pPbukZN8Ul2vCRRVdZ02+RkN5mnYiUiL44BcezCyoWf4vwCuRSCuy8FMSVa38=" -sourcefile test -targetfile /tmp/from25

// ./main parse -path=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -from=2 -to=-1 -output=output/ -remotepath=/root/testauto -transfer=true -username=root -password=***REMOVED*** -watch=true -interval=1
//go run main.go parse -inputpath=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -rowstart=2 -rowend=-1 -outputpath=output/ -outputtype=txt -remotepath=/root/testauto -transfer=tr> go run main.go parse -inputpath=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -rowstart=2 -rowend=-1 -outputpath=output/ -outputtype=txt -remotepath=/root/testauto -transfer=true -username=root -password=***REMOVED*** -interval=1 -hostkey="***REMOVED***"
package main

import (
	"excel"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var watcher *fsnotify.Watcher

func usage() {
	fmt.Println("Usage: parse_excel <command> [<args>]")
	fmt.Println()
	fmt.Println("Availabve commands are: ")
	fmt.Println("    parse: Parse the excel directory.")
}

func main() {
	if len(os.Args) == 1 {
		usage()
		return
	}

	parseCommand := flag.NewFlagSet("parse", flag.ExitOnError) // {{{
	inputPath := parseCommand.String(
		"inputpath",
		"test.xlsx",
		"filename or directory to parse",
	)
	sheetIndex := parseCommand.Int(
		"sheet",
		0,
		"sheet index",
	)
	rowStartsAt := parseCommand.Int(
		"rowstart",
		0,
		"first row index of the range",
	)
	rowEndsAt := parseCommand.Int(
		"rowend",
		-1,
		"last row index of the range, '-1' means all the rest rows",
	)
	columnIndices := parseCommand.String(
		"columns",
		"0,3,5,7,11",
		"colmuns to be extracted",
	)
	outputType := parseCommand.String(
		"outputtype",
		"xlsx",
		"type of output file, csv, txt or xlsx",
	)
	outputPath := parseCommand.String(
		"outputpath",
		"",
		"path of output directory",
	)
	remoteOutputPath := parseCommand.String(
		"remotepath",
		"",
		"path of output directory",
	)
	isTransfer := parseCommand.Bool(
		"transfer",
		false,
		"enable output file transfer",
	)
	hostKey := parseCommand.String(
		"hostkey",
		"101.201.180.67 ecdsa-sha2-nistp256 xxx",
		"lines in ./ssh/known_host",
	)
	userName := parseCommand.String(
		"username",
		"root",
		"user name to use when connecting to remote server",
	)
	password := parseCommand.String(
		"password",
		"",
		"password to use when connecting to remote server",
	)
	isWatch := parseCommand.Bool(
		"watch",
		false,
		"enable to watch the directory",
	)
	interval := parseCommand.Int(
		"interval",
		60,
		"interval of walking through the folders, not for files",
	)
	logLevel := parseCommand.String(
		"loglevel",
		"info",
		"log level",
	)
	// }}}

	switch os.Args[1] {
	case "parse":
		parseCommand.Parse(os.Args[2:])
	case "-h":
		usage()
		return
	default:
		fmt.Printf("invalid command: %q", os.Args[1])
		usage()
		return
	}

	if parseCommand.Parsed() {
		columns, err := ParseColumnIndices(*columnIndices)
		if err != nil {
			fmt.Println(err)
			return
		}
		isFolder, err := IsFolder(*inputPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		logFile := filepath.Join(*outputPath, "log.txt")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("failed to open log file:", logFile)
			return
		}
		mw := io.MultiWriter(os.Stdout, file)
		logrus.SetOutput(mw)
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:          true,
			DisableLevelTruncation: false,
		})
		switch *logLevel {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
		logrus.WithFields(logrus.Fields{
			"logfile":  logFile,
			"loglevel": *logLevel,
		}).Info("LOG")
		logrus.WithFields(logrus.Fields{
			"inputpath":  *inputPath,
			"sheet":      *sheetIndex,
			"rowstart":   *rowStartsAt,
			"rowend":     *rowEndsAt,
			"columns":    *columnIndices,
			"outputpath": *outputPath,
			"outputtype": *outputType,
			"hostkey":    *hostKey,
			"username":   *userName,
			"password":   *password,
			"remotepath": *remoteOutputPath,
			"transfer":   *isTransfer,
			"watch":      *isWatch,
			"interval":   *interval,
		}).Debug("LOG")

		if !isFolder {
			if outputFile, err := Extract(
				*inputPath,
				*sheetIndex,
				*rowStartsAt,
				*rowEndsAt,
				columns,
				*outputPath,
				*outputType,
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
		if !*isWatch {
			// TODO: parse all the files and transfer
			if err := filepath.Walk(
				*inputPath,
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
					if outputFile, err := Extract(
						inputPath,
						*sheetIndex,
						*rowStartsAt,
						*rowEndsAt,
						columns,
						*outputPath,
						*outputType,
					); err != nil {
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": err.Error(),
						}).Error("PRS")
					} else {
						logrus.WithFields(logrus.Fields{
							"file": outputFile,
						}).Info("PRS")
						if !*isTransfer {
							return nil
						}
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": "start",
						}).Debug("SND")
						if err := Send(
							*hostKey,
							*userName,
							*password,
							inputPath,
							outputFile,
							*outputType,
							*remoteOutputPath,
						); err != nil {
							logrus.WithFields(logrus.Fields{
								"file":    inputPath,
								"message": err.Error(),
							}).Error("SND")
						} else {
							logrus.WithFields(logrus.Fields{
								"file":    inputPath,
								"message": "sent",
							}).Info("SND")
						}
					}
					return nil
				}); err != nil {
				logrus.WithFields(logrus.Fields{
					"path":    inputPath,
					"message": err.Error(),
				}).Error("ADD")
			}
			return
		}
		watcher, _ = fsnotify.NewWatcher()
		defer watcher.Close()
		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
						logrus.WithFields(logrus.Fields{
							"file": event.Name,
						}).Info("NEW")
						if outputFile, err := Extract(
							event.Name,
							*sheetIndex,
							*rowStartsAt,
							*rowEndsAt,
							columns,
							*outputPath,
							*outputType,
						); err != nil {
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": err.Error(),
							}).Error("PRS")
						} else {
							logrus.WithFields(logrus.Fields{
								"file": outputFile,
							}).Info("PRS")
							if !*isTransfer {
								continue
							}
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": "start",
							}).Debug("SND")
							if err := Send(
								*hostKey,
								*userName,
								*password,
								event.Name,
								outputFile,
								*outputType,
								*remoteOutputPath,
							); err != nil {
								logrus.WithFields(logrus.Fields{
									"file":    event.Name,
									"message": err.Error(),
								}).Error("SND")
							} else {
								logrus.WithFields(logrus.Fields{
									"file":    event.Name,
									"message": "sent",
								}).Info("SND")
							}
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					logrus.WithFields(logrus.Fields{
						"message": err.Error(),
					}).Error("LOG")
				}
			}
		}()
		go Watch(*inputPath, *interval)
		<-done
	}
}

func IsFolder(inputpath string) (isFolder bool, err error) { // {{{
	f, err := os.Stat(inputpath)
	if err != nil {
		return isFolder, err
	}
	return f.Mode().IsDir(), nil
} // }}}

func Watch(inputPath string, duration int) { // {{{
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(time.Duration(duration) * time.Second)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		<-done
		if err := filepath.Walk(inputPath, addToWatcher); err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    inputPath,
				"message": err.Error(),
			}).Error("ADD")
		}
		go func() {
			done <- struct{}{}
		}()
	}
} // }}}

func addToWatcher(inputPath string, f os.FileInfo, err error) error { // {{{
	if f.Mode().IsDir() {
		logrus.WithFields(logrus.Fields{
			"file": inputPath,
		}).Debug("ADD")
		return watcher.Add(inputPath)
	}
	return nil
} // }}}

func NewFileName(outputPath, name, ext string) string { // {{{
	dir, file := filepath.Split(name)
	if outputPath != "_" {
		dir = outputPath
	}
	return filepath.Join(
		dir,
		fmt.Sprintf(
			"%v.%v",
			strings.TrimSuffix(file, path.Ext(file)),
			ext,
		),
	)
} // }}}

func ParseColumnIndices(columnIndices string) (output []int, err error) { // {{{
	for _, indexs := range strings.Split(columnIndices, ",") {
		indexi, err := strconv.Atoi(indexs)
		if err != nil {
			return output, fmt.Errorf(
				"invalid integer '%v' in '%v'",
				indexs,
				columnIndices,
			)
		}
		output = append(output, indexi)
	}
	return output, nil
} // }}}

func Extract( // {{{
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
	outputFile = NewFileName(
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
	return outputFile, nil
} // }}}

func Send( // {{{
	hostKey string,
	username string,
	password string,
	sourceFile string,
	outputFile string,
	outputType string,
	remoteOutputPath string,
) error {
	if username == "" {
		return fmt.Errorf("missing username")
	}
	if password == "" {
		return fmt.Errorf("missing password")
	}
	if sourceFile == "" {
		return fmt.Errorf("missing sourcefile")
	}

	_, hosts, pubKey, _, _, err := ssh.ParseKnownHosts([]byte(hostKey))
	if err != nil {
		return fmt.Errorf("invalid host key: %v", err)
	}
	if len(hosts) < 1 {
		return fmt.Errorf("invalid host: %v", hosts)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.FixedHostKey(pubKey),
	}
	conn, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:22", hosts[0]),
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	remoteDir, remoteFile := MakeRemoteFile(sourceFile, remoteOutputPath, outputType)
	logrus.WithFields(logrus.Fields{
		"outputFile": outputFile,
		"remoteDir":  remoteDir,
		"remoteFile": remoteFile,
	}).Debug("SND")

	if remoteDir != "" && remoteDir != "./" && remoteDir != "~/" {
		if client.MkdirAll(remoteDir) != nil {
			return fmt.Errorf(
				"failed to create remote directory '%s': %v",
				remoteDir,
				err,
			)
		}
	}
	dstFile, err := client.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	defer dstFile.Close()

	srcFile, err := os.Open(outputFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}
	logrus.WithFields(logrus.Fields{
		"bytesSent": bytes,
	}).Debug("SND")

	return nil
} // }}}

func MakeRemoteFile(sourceFile string, remoteOutputPath string, outputType string) (string, string) { // {{{
	sourceDir, sourceFile := filepath.Split(sourceFile)
	monthDir := path.Base(sourceDir)
	remoteDir := filepath.Join(remoteOutputPath, monthDir)
	sourceFileName := strings.TrimSuffix(sourceFile, path.Ext(sourceFile))
	remoteFileName := fmt.Sprintf(
		"%v.%v",
		sourceFileName,
		outputType,
	)
	remoteDir = filepath.Join(remoteDir, sourceFileName)
	remoteFile := filepath.Join(remoteDir, remoteFileName)
	return remoteDir, remoteFile
} // }}}
