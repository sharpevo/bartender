package fsop

import (
	"fmt"
	"path"
	"path/filepath"
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
