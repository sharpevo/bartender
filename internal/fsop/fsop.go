package fsop

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func MakeRemoteFileWithinNameFolder(
	sourceFilePath string,
	remoteOutputPath string,
	outputType string,
) (string, string) { // {{{
	sourceDir, sourceFile := filepath.Split(sourceFilePath)
	monthDir := path.Base(sourceDir)
	remoteDir := filepath.Join(remoteOutputPath, monthDir)
	sourceFileName := strings.TrimSuffix(sourceFile, path.Ext(sourceFile))
	remoteDir = filepath.Join(remoteDir, sourceFileName)
	remoteFileName := fmt.Sprintf(
		"%v.%v",
		sourceFileName,
		outputType,
	)
	remoteFile := filepath.Join(remoteDir, remoteFileName)
	return remoteDir, remoteFile
} // }}}

func CustomRemoteFileNameAndDir( // {{{
	sourceFilePath string,
	remoteOutputPath string,
	outputType string,
) (string, string) {
	sourceDir, sourceFile := filepath.Split(sourceFilePath)
	monthDir := path.Base(sourceDir)
	remoteDir := filepath.Join(remoteOutputPath, monthDir)
	sourceFileName := strings.TrimSuffix(sourceFile, path.Ext(sourceFile))
	remoteDir = filepath.Join(remoteDir, sourceFileName)
	remoteFileName := fmt.Sprintf(
		"%v.%v",
		sourceFileName,
		outputType,
	)
	return remoteDir, remoteFileName
} // }}}

func IsDir(inputpath string) (isFolder bool, err error) { // {{{
	f, err := os.Stat(inputpath)
	if err != nil {
		return isFolder, err
	}
	return f.Mode().IsDir(), nil
} // }}}

func ConvertColumnIndices(columnIndices string) (output []int, err error) { // {{{
	for _, indexs := range strings.Split(columnIndices, ",") {
		indexi, err := strconv.Atoi(strings.Trim(indexs, " "))
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

func MakeOutputFilePath(outputPath string, fileName string, ext string) string { // {{{
	_, file := filepath.Split(fileName)
	return filepath.Join(
		outputPath,
		fmt.Sprintf(
			"%v.%v",
			strings.TrimSuffix(file, path.Ext(file)),
			ext,
		),
	)
} // }}}
