package fsop_test

import (
	"automation/parser/internal/fsop"
	"fmt"
	"testing"
)

// remote folder is composed with the folder of source file and the file name.
func TestMakeRemoteFileWithinNameFolder(t *testing.T) {
	cases := []struct {
		source     string
		remote     string
		outputtype string
		expectDir  string
		expectFile string
	}{
		{
			source:     "/var/log/test/project-A.xlsm",
			remote:     "/root/upload",
			outputtype: "txt",
			expectDir:  "/root/upload/test/project-A",
			expectFile: "/root/upload/test/project-A/project-A.txt",
		},
	}

	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			dir, name := fsop.MakeRemoteFileWithinNameFolder(c.source, c.remote, c.outputtype)
			fmt.Println(dir, name)
			if dir != c.expectDir || name != c.expectFile {
				t.Errorf(
					"\nEXPECT: %v %v\n GET: %v %v\n\n",
					c.expectDir,
					c.expectFile,
					dir,
					name,
				)
			}
		})
	}
}
