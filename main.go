//
// parse -file ../20190722-Nova314--分析需求表-贝瑞\(1\).xlsm -path /tmp/ -from 2 -sheet 1 -to -1 -type csv
// watch -directory /tmp/watcher -from 2 -sheet 1 -to -1 -type csv
// send -username wuy -password igenetech -hostkey "192.168.1.96 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBLjYGzkYWF+a1KV2NDjEtjzfa0pPbukZN8Ul2vCRRVdZ02+RkN5mnYiUiL44BcezCyoWf4vwCuRSCuy8FMSVa38=" -sourcefile test -targetfile /tmp/from25

// ./main parse -path=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -from=2 -to=-1 -output=output/ -remotepath=/root/testauto -transfer=true -username=root -password=***REMOVED*** -watch=true -interval=1
//go run main.go parse -inputpath=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -rowstart=2 -rowend=-1 -outputpath=output/ -outputtype=txt -remotepath=/root/testauto -transfer=tr> go run main.go parse -inputpath=/tmp/watcher -sheet=1 -columns=1,2,3,4,5,9,11 -rowstart=2 -rowend=-1 -outputpath=output/ -outputtype=txt -remotepath=/root/testauto -transfer=true -username=root -password=***REMOVED*** -interval=1 -hostkey="***REMOVED***"
package main

import (
	"automation/parser/internal/fsop"
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
	"path/filepath"
	"regexp"
	"time"
)

type ParseOptions struct { // {{{
	InputPath     string
	SheetIndex    int
	RowStartsAt   int
	RowEndsAt     int
	ColumnIndices string
	OutputType    string
	OutputPath    string
}

func AttachParseOptions(cmd *flag.FlagSet) *ParseOptions {
	options := &ParseOptions{}
	cmd.StringVar(
		&options.InputPath,
		"inputpath",
		"test.xlsx",
		"filename or directory to parse",
	)
	cmd.IntVar(
		&options.SheetIndex,
		"sheet",
		0,
		"sheet index",
	)
	cmd.IntVar(
		&options.RowStartsAt,
		"rowstart",
		0,
		"first row index of the range",
	)
	cmd.IntVar(
		&options.RowEndsAt,
		"rowend",
		-1,
		"last row index of the range, '-1' means all the rest rows",
	)
	cmd.StringVar(
		&options.ColumnIndices,
		"columns",
		"0,3,5,7,11",
		"colmuns to be extracted",
	)
	cmd.StringVar(
		&options.OutputType,
		"outputtype",
		"xlsx",
		"type of output file, csv, txt or xlsx",
	)
	cmd.StringVar(
		&options.OutputPath,
		"outputpath",
		"",
		"path of output directory",
	)
	return options
} // }}}

type TransferOptions struct { // {{{
	InputPath string
}

func AttachTransferOptions(cmd *flag.FlagSet) *TransferOptions {
	options := &TransferOptions{}
	cmd.StringVar(
		&options.InputPath,
		"sourcepath",
		"",
		"local directory to transfer",
	)
	return options
} // }}}

type ServerOptions struct { // {{{
	Enabled   bool
	HostKey   string
	UserName  string
	Password  string
	Directory string
}

func AttachServerOptions(cmd *flag.FlagSet) *ServerOptions {
	options := &ServerOptions{}
	cmd.BoolVar(
		&options.Enabled,
		"transfer",
		false,
		"enable output file transfer",
	)
	cmd.StringVar(
		&options.HostKey,
		"hostkey",
		"101.201.180.67 ecdsa-sha2-nistp256 xxx",
		"lines in ./ssh/known_host",
	)
	cmd.StringVar(
		&options.UserName,
		"username",
		"root",
		"user name to use when connecting to remote server",
	)
	cmd.StringVar(
		&options.Password,
		"password",
		"",
		"password to use when connecting to remote server",
	)
	cmd.StringVar(
		&options.Directory,
		"remotepath",
		"",
		"path of output directory",
	)
	return options
} // }}}

type WatchOptions struct { // {{{
	Enabled         bool
	Interval        time.Duration
	FileNamePattern string
}

func AttachWatchOptions(cmd *flag.FlagSet) *WatchOptions {
	options := &WatchOptions{}
	cmd.BoolVar(
		&options.Enabled,
		"watch",
		false,
		"enable to watch the directory",
	)
	cmd.DurationVar(
		&options.Interval,
		"interval",
		10*time.Second,
		"interval of walking through the folders, not for files",
	)
	cmd.StringVar(
		&options.FileNamePattern,
		"namepattern",
		"",
		"filename pattern",
	)
	return options
} // }}}

type LogOptions struct { // {{{
	Level string
}

func AttachLogOptions(cmd *flag.FlagSet) *LogOptions {
	options := &LogOptions{}
	cmd.StringVar(
		&options.Level,
		"loglevel",
		"info",
		"log level",
	)
	return options
} // }}}

var watcher *fsnotify.Watcher

func init() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("failed to open log file 'log.txt'")
		return
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logrus.SetOutput(mw)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:          true,
		DisableLevelTruncation: false,
	})

}
func usage() {
	fmt.Println("Usage: parse_excel <command> [<args>]")
	fmt.Println()
	fmt.Println("Availabve commands are: ")
	fmt.Println("    parse: Parse the excel directory.")
	fmt.Println("    transfer: transfer files.")
}

func main() {
	if len(os.Args) == 1 {
		usage()
		return
	}

	parseCommand := flag.NewFlagSet("parse", flag.ExitOnError)
	parseParseOptions := AttachParseOptions(parseCommand)
	parseServerOptions := AttachServerOptions(parseCommand)
	parseWatchOptions := AttachWatchOptions(parseCommand)
	parseLogOptions := AttachLogOptions(parseCommand)

	transferCommand := flag.NewFlagSet("transfer", flag.ExitOnError)
	transferTransferOptions := AttachTransferOptions(transferCommand)
	transferServerOptions := AttachServerOptions(transferCommand)
	transferWatchOptions := AttachWatchOptions(transferCommand)
	//transferLogOptions := AttachLogOptions(transferCommand)

	switch os.Args[1] {
	case "parse":
		parseCommand.Parse(os.Args[2:])
	case "transfer":
		transferCommand.Parse(os.Args[2:])
	case "-h":
		usage()
		return
	default:
		fmt.Printf("invalid command: %q", os.Args[1])
		usage()
		return
	}

	if parseCommand.Parsed() { // {{{
		columns, err := fsop.ConvertColumnIndices(parseParseOptions.ColumnIndices)
		if err != nil {
			fmt.Println(err)
			return
		}
		isFolder, err := fsop.IsDir(parseParseOptions.InputPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch parseLogOptions.Level {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
		logrus.WithFields(logrus.Fields{
			"loglevel": parseLogOptions.Level,
		}).Info("LOG")
		logrus.WithFields(logrus.Fields{
			"inputpath":  parseParseOptions.InputPath,
			"sheet":      parseParseOptions.SheetIndex,
			"rowstart":   parseParseOptions.RowStartsAt,
			"rowend":     parseParseOptions.RowEndsAt,
			"columns":    parseParseOptions.ColumnIndices,
			"outputpath": parseParseOptions.OutputPath,
			"outputtype": parseParseOptions.OutputPath,
			"hostkey":    parseServerOptions.HostKey,
			"username":   parseServerOptions.UserName,
			"password":   parseServerOptions.Password,
			"remotepath": parseServerOptions.Directory,
			"transfer":   parseServerOptions.Enabled,
			"watch":      parseWatchOptions.Enabled,
			"interval":   parseWatchOptions.Interval,
		}).Debug("LOG")

		if !isFolder {
			if outputFile, err := Extract(
				parseParseOptions.InputPath,
				parseParseOptions.SheetIndex,
				parseParseOptions.RowStartsAt,
				parseParseOptions.RowEndsAt,
				columns,
				parseParseOptions.OutputPath,
				parseParseOptions.OutputType,
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
		if !parseWatchOptions.Enabled { // {{{
			if err := filepath.Walk(
				parseParseOptions.InputPath,
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
						parseParseOptions,
						parseServerOptions,
					) != nil {
						logrus.WithFields(logrus.Fields{
							"file":    inputPath,
							"message": err.Error(),
						}).Error("PRS")
					}
					return nil
				}); err != nil {
				logrus.WithFields(logrus.Fields{
					"path":    parseParseOptions.InputPath,
					"message": err.Error(),
				}).Error("PRS")
			}
			return
		} // }}}
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
						if HandleParse(
							event.Name,
							columns,
							parseParseOptions,
							parseServerOptions,
						) != nil {
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": err.Error(),
							}).Error("PRS")
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
		go Watch(parseParseOptions.InputPath, parseWatchOptions.Interval)
		<-done
	} // }}}
	if transferCommand.Parsed() { // {{{
		r, err := regexp.Compile(transferWatchOptions.FileNamePattern)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path":    transferTransferOptions.InputPath,
				"message": err.Error(),
			}).Error("TRS")
		}
		if !transferWatchOptions.Enabled {
			if err := filepath.Walk(
				transferTransferOptions.InputPath,
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
					if HandleTransfer(inputPath, transferServerOptions) != nil {
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
					"path":    transferTransferOptions.InputPath,
					"message": err.Error(),
				}).Error("TRS")
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
						if !r.MatchString(event.Name) {
							continue
						}
						if HandleTransfer(event.Name, transferServerOptions) != nil {
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": err.Error(),
							}).Error("TRS")
						} else {
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": "sent",
							}).Info("TRS")
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
		go Watch(transferTransferOptions.InputPath, transferWatchOptions.Interval)
		<-done
	} // }}}
}

func Watch(inputPath string, duration time.Duration) { // {{{
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for ; ; <-ticker.C {
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
	return outputFile, nil
} // }}}

func Send( // {{{
	hostKey string,
	username string,
	password string,
	localFilepath string,
	remoteFilename string,
	remoteDir string,
) error {
	if username == "" {
		return fmt.Errorf("missing username")
	}
	if password == "" {
		return fmt.Errorf("missing password")
	}
	if localFilepath == "" {
		return fmt.Errorf("missing localFilepath")
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

	remoteFilepath := filepath.Join(remoteDir, remoteFilename)
	if remoteDir != "" && remoteDir != "./" && remoteDir != "~/" {
		if client.MkdirAll(remoteDir) != nil {
			return fmt.Errorf(
				"failed to create remote directory '%s': %v",
				remoteDir,
				err,
			)
		}
	}
	dstFile, err := client.Create(remoteFilepath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	if client.Chmod(remoteFilepath, os.FileMode(0755)) != nil {
		logrus.WithFields(logrus.Fields{
			"file":    remoteFilepath,
			"message": "failed to chmod",
		}).Error("SND")
	}
	defer dstFile.Close()

	srcFile, err := os.Open(localFilepath)
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

// TODO: log to extract and send
func HandleParse( // {{{
	inputPath string,
	columns []int,
	parseParseOptions *ParseOptions,
	parseServerOptions *ServerOptions,
) error {
	outputFile, err := Extract(
		inputPath,
		parseParseOptions.SheetIndex,
		parseParseOptions.RowStartsAt,
		parseParseOptions.RowEndsAt,
		columns,
		parseParseOptions.OutputPath,
		parseParseOptions.OutputType,
	)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"file": outputFile,
	}).Info("PRS")
	if !parseServerOptions.Enabled {
		return nil
	}
	logrus.WithFields(logrus.Fields{
		"file":    inputPath,
		"message": "start",
	}).Debug("SND")

	remoteDir, remoteFileName := fsop.CustomRemoteFileNameAndDir(
		inputPath,
		parseServerOptions.Directory,
		parseParseOptions.OutputType,
	)
	logrus.WithFields(logrus.Fields{
		"outputFile":     outputFile,
		"remoteDir":      remoteDir,
		"remoteFileName": remoteFileName,
	}).Debug("SND")
	if Send(
		parseServerOptions.HostKey,
		parseServerOptions.UserName,
		parseServerOptions.Password,
		outputFile,
		remoteFileName,
		remoteDir,
	) != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"local":  outputFile,
		"remote": remoteDir,
	}).Info("SND")
	return nil
} // }}}

func HandleTransfer(
	inputPath string,
	transferServerOptions *ServerOptions,
) error {
	if !transferServerOptions.Enabled {
		return nil
	}
	_, fileName := filepath.Split(inputPath)
	return Send(
		transferServerOptions.HostKey,
		transferServerOptions.UserName,
		transferServerOptions.Password,
		inputPath,
		fileName,
		transferServerOptions.Directory,
	)
}
