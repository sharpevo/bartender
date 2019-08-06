//
// parse -file ../20190722-Nova314--分析需求表-贝瑞\(1\).xlsm -path /tmp/ -from 2 -sheet 1 -to -1 -type csv
// watch -directory /tmp/watcher -from 2 -sheet 1 -to -1 -type csv
// send -username wuy -password igenetech -hostkey "192.168.1.96 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBLjYGzkYWF+a1KV2NDjEtjzfa0pPbukZN8Ul2vCRRVdZ02+RkN5mnYiUiL44BcezCyoWf4vwCuRSCuy8FMSVa38=" -sourcefile test -targetfile /tmp/from25
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
)

const (
	WATCHER_DIRECTORY = "watcher"
)

func usage() {
	fmt.Println("Usage: parse_excel <command> [<args>]")
	fmt.Println()
	fmt.Println("Availabve commands are: ")
	fmt.Println("    parse: Parse the excel directory.")
	fmt.Println("    send: Send file to remote directory.")
	fmt.Println("    watch: Watch a directory and parse the excel when created.")
}

func main() {
	if len(os.Args) == 1 {
		usage()
		return
	}

	parseCommand := flag.NewFlagSet("parse", flag.ExitOnError)
	pfileName := parseCommand.String(
		"file",
		"test.xlsx",
		"",
	)
	psheetIndex := parseCommand.Int(
		"sheet",
		1,
		"sheet index",
	)
	prowStartsAt := parseCommand.Int(
		"from",
		2,
		"first row index of the range",
	)
	prowEndsAt := parseCommand.Int(
		"to",
		-1,
		"last row index of the range, '-1' means all the rest rows",
	)
	pcolumnIndices := parseCommand.String(
		"columns",
		"1,2,3,4,5,9,11",
		"colmuns to be extracted",
	)
	poutputType := parseCommand.String(
		"type",
		"csv",
		"type of output file, csv or xlsx",
	)
	poutputPath := parseCommand.String(
		"path",
		"_",
		"path of output file, '_' means same folder of the input file",
	)

	sendCommand := flag.NewFlagSet("send", flag.ExitOnError)
	hostKey := sendCommand.String(
		"hostkey",
		"***REMOVED***",
		"lines in ./ssh/known_host",
	)
	userName := sendCommand.String(
		"username",
		"root",
		"user name to use when connecting to remote server",
	)
	password := sendCommand.String(
		"password",
		"",
		"password to use when connecting to remote server",
	)
	sourceFile := sendCommand.String(
		"sourcefile",
		"",
		"the source file to send to remote server",
	)
	targetFile := sendCommand.String(
		"targetfile",
		"",
		"the target file to send to remote server",
	)

	watchCommand := flag.NewFlagSet("watch", flag.ExitOnError)
	watchPath := watchCommand.String(
		"directory",
		"/tmp/watch",
		"the path of directory to watch",
	)
	wsheetIndexp := watchCommand.Int(
		"sheet",
		1,
		"sheet index",
	)
	wrowStartsAtp := watchCommand.Int(
		"from",
		2,
		"first row index of the range",
	)
	wrowEndsAtp := watchCommand.Int(
		"to",
		-1,
		"last row index of the range, '-1' means all the rest rows",
	)
	wcolumnIndices := watchCommand.String(
		"columns",
		"0,1,2,3,4,9,11",
		"colmuns to be extracted",
	)
	woutputType := watchCommand.String(
		"type",
		"csv",
		"type of output file, csv or xlsx",
	)
	dryRun := watchCommand.Bool(
		"dry",
		false,
		"perform a trial run with no data sent",
	)
	whostKey := watchCommand.String(
		"hostkey",
		"***REMOVED***",
		"lines in ./ssh/known_host",
	)
	wuserName := watchCommand.String(
		"username",
		"root",
		"user name to use when connecting to remote server",
	)
	wpassword := watchCommand.String(
		"password",
		"",
		"password to use when connecting to remote server",
	)
	wsourceFile := watchCommand.String(
		"sourcefile",
		"",
		"the source file to send to remote server",
	)
	wtargetFile := watchCommand.String(
		"targetfile",
		"",
		"the target file to send to remote server",
	)

	switch os.Args[1] {
	case "parse":
		parseCommand.Parse(os.Args[2:])
	case "send":
		sendCommand.Parse(os.Args[2:])
	case "watch":
		watchCommand.Parse(os.Args[2:])
	case "-h":
		usage()
	default:
		fmt.Printf("invalid command: %q", os.Args[1])
	}

	if watchCommand.Parsed() {

		logFile := filepath.Join(WATCHER_DIRECTORY, "log.txt")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			fmt.Println("failed to open log file:", logFile)
			return
		}
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
		})
		logrus.SetLevel(logrus.TraceLevel)
		columnIndices, err := ParseColumnIndices(*wcolumnIndices)
		if err != nil {
			fmt.Println(err)
			return
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Println(err)
			return
		}
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
							*wsheetIndexp,
							*wrowStartsAtp,
							*wrowEndsAtp,
							columnIndices,
							WATCHER_DIRECTORY,
							*woutputType,
						); err != nil {
							logrus.WithFields(logrus.Fields{
								"file":    event.Name,
								"message": err.Error(),
							}).Error("PRC")
						} else {
							log.Println(outputFile)
							logrus.WithFields(logrus.Fields{
								"file": outputFile,
							}).Info("PRC")
						}
						if !*dryRun {
							if err := Send(
								*whostKey,
								*wuserName,
								*wpassword,
								*wsourceFile,
								*wtargetFile,
							); err != nil {
								log.Println(err)
							}
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					fmt.Println("error:", err)
				}
			}
		}()
		err = watcher.Add(*watchPath)
		if err != nil {
			fmt.Println(err)
		}
		<-done
	}

	if parseCommand.Parsed() {
		columnIndices, err := ParseColumnIndices(*pcolumnIndices)
		if err != nil {
			fmt.Println(err)
			return
		}
		if outputFile, err := Extract(
			*pfileName,
			*psheetIndex,
			*prowStartsAt,
			*prowEndsAt,
			columnIndices,
			*poutputPath,
			*poutputType,
		); err != nil {
			log.Println(err)
			return
		} else {
			log.Println(outputFile)
			return
		}
	}

	if sendCommand.Parsed() {
		if err := Send(
			*hostKey,
			*userName,
			*password,
			*sourceFile,
			*targetFile,
		); err != nil {
			log.Println(err)
		}
	}

}

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
	case "csv":
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

func Send(
	hostKey string,
	username string,
	password string,
	sourceFile string,
	targetFile string,
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
	if targetFile == "" {
		return fmt.Errorf("missing targetfile")
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

	dir, _ := filepath.Split(targetFile)
	if dir != "" && dir != "./" {
		if client.MkdirAll(dir) != nil {
			return fmt.Errorf(
				"failed to create target directory '%s': %v",
				dir,
				err,
			)
		}
	}
	dstFile, err := client.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	defer dstFile.Close()

	srcFile, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}
	fmt.Printf("%d bytes copied\n", bytes)
	return nil
}
