package fsop_test

import (
	"automation/parser/internal/fsop"
	"fmt"
	"reflect"
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

func TestIsDir(t *testing.T) {
	cases := []struct {
		input       string
		expectBool  bool
		expectError string
	}{
		{
			input:       "/var/log",
			expectBool:  true,
			expectError: "",
		},
		{
			input:       "/var/log/nonexist",
			expectBool:  false,
			expectError: "stat /var/log/nonexist: no such file or directory",
		},
	}
	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			isDir, err := fsop.IsDir(c.input)
			if isDir != c.expectBool ||
				(err != nil && err.Error() != c.expectError) {
				t.Errorf(
					"\nEXPECT: %v %v\n GET: %v %v\n\n",
					c.expectBool,
					c.expectError,
					isDir,
					err,
				)
			}
		})
	}
}

func TestConvertColumnIndices(t *testing.T) {
	cases := []struct {
		columnIndices string
		expectOutput  []int
		expectError   string
	}{
		{
			columnIndices: "0,1,2,7",
			expectOutput:  []int{0, 1, 2, 7},
			expectError:   "",
		},
		{
			columnIndices: "2, 4, 6",
			expectOutput:  []int{2, 4, 6},
			expectError:   "",
		},
	}
	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			output, err := fsop.ConvertColumnIndices(c.columnIndices)
			if !reflect.DeepEqual(output, c.expectOutput) ||
				(err != nil && err.Error() != c.expectError) {
				t.Errorf(
					"\nEXPECT: %v %v\n GET: %v %v\n\n",
					c.expectOutput,
					c.expectError,
					output,
					err,
				)
			}
		})
	}
}
