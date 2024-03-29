package fsop_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sharpevo/bartender/internal/fsop"
)

func TestCustomRemoteFileNameAndDir(t *testing.T) {
	cases := []struct {
		source     string
		remote     string
		outputtype string
		expectDir  string
		expectFile string
	}{
		{
			source:     "/var/www/html/cloudrun/chart/202103/20200222-NSS40153--分析需求表-艾吉.xlsm",
			remote:     "/root/upload",
			outputtype: "txt",
			expectDir:  "/root/upload/202002/20200222-NSS40153--分析需求表-艾吉",
			expectFile: "20200222-NSS40153--分析需求表-艾吉.txt",
		},
		{
			source:     "/var/log/test/project-A.xlsm",
			remote:     "/root/upload",
			outputtype: "txt",
			//expectDir:  "/root/upload/test/project-A",
			expectDir:  "/root/upload/projec/project-A",
			expectFile: "project-A.txt",
		},
	}

	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			dir, name := fsop.CustomRemoteFileNameAndDir(c.source, c.remote, c.outputtype)
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

func TestMakeOutputFilePath(t *testing.T) {
	cases := []struct {
		outputpath string
		inputpath  string
		fileext    string
		expect     string
	}{
		{
			outputpath: "/tmp/output",
			inputpath:  "testfile.xlsx",
			fileext:    "csv",
			expect:     "/tmp/output/testfile.csv",
		},
		{
			outputpath: "",
			inputpath:  "testfile.xlsx",
			fileext:    "txt",
			expect:     "testfile.txt",
		},
		{
			outputpath: "/tmp/output",
			inputpath:  "/a/b/c/testfile.xlsx",
			fileext:    "csv",
			expect:     "/tmp/output/a/b/c/testfile.csv",
		},
	}
	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			outputfilepath := fsop.MakeOutputFilePath(
				c.outputpath, c.inputpath, c.fileext)
			if outputfilepath != c.expect {
				t.Errorf(
					"\nEXPECT: %v\n GET: %v\n\n",
					c.expect,
					outputfilepath,
				)
			}
		})
	}
}

func TestGetRelativePath(t *testing.T) {
	cases := []struct {
		base   string
		input  string
		expect string
	}{
		{
			base:   "/a/b/",
			input:  "/a/b/c/d/e.txt",
			expect: "c/d/",
		},
		{
			base:   "/a/b",
			input:  "/a/b/c/d/e.txt",
			expect: "c/d/",
		},
		{
			base:   "~/a/b",
			input:  "/home/yang/a/b/c/d/e.txt",
			expect: "c/d/",
		},
		{
			base:   "a/b",
			input:  "a/b/c.txt",
			expect: "",
		},
		{
			base:   "a/b",
			input:  "a/b/c/d/e.txt",
			expect: "c/d/",
		},
	}
	for index, c := range cases {
		t.Run(fmt.Sprintf("%v", index), func(t *testing.T) {
			actual, _ := fsop.GetRelativePath(c.base, c.input)
			if actual != c.expect {
				t.Errorf(
					"\nEXPECT: %v\n GET: %v\n\n",
					c.expect,
					actual,
				)
			}
		})
	}
}
