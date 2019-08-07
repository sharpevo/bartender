package fsop

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func MakeRemoteFileWithinNameFolder(sourceFilePath string, remoteOutputPath string, outputType string) (string, string) { // {{{
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
}
