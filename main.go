package main

import (
	"excel"
	"flag"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	fileNamep := flag.String(
		"file",
		"test.xlsx",
		"",
	)
	sheetIndexp := flag.Int(
		"sheet",
		1,
		"sheet index",
	)
	rowStartsAtp := flag.Int(
		"from",
		2,
		"first row index of the range",
	)
	rowEndsAtp := flag.Int(
		"to",
		-1,
		"last row index of the range, '-1' means all the rest rows",
	)
	columnIndicesp := flag.String(
		"columns",
		"0,1,2,3,4,9,11",
		"colmuns to be extracted",
	)
	outputType := flag.String(
		"type",
		"csv",
		"type of output file, csv or xlsx",
	)
	outputPath := flag.String(
		"path",
		"_",
		"path of output file, '_' means same folder of the input file",
	)
	flag.Parse()
	columnIndices := []int{}
	for _, indexs := range strings.Split(*columnIndicesp, ",") {
		indexi, err := strconv.Atoi(indexs)
		if err != nil {
			fmt.Printf(
				"invalid integer '%v' in '%v'",
				indexs,
				*columnIndicesp,
			)
			return
		}
		columnIndices = append(columnIndices, indexi)
	}
	data, err := excel.ExtractColumns(
		*fileNamep,
		*sheetIndexp,
		*rowStartsAtp,
		*rowEndsAtp,
		columnIndices,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	outputFile := NewFileName(*outputPath, *fileNamep, *outputType)
	switch *outputType {
	case "csv":
		if excel.MakeFileCSV(
			outputFile,
			data,
			excel.SEPARATOR_TAB,
		) != nil {
			fmt.Println(err)
			return
		}
	case "xlsx":
		if excel.MakeFileXLSX(
			outputFile,
			data,
			"sheet-0",
		) != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Printf(
			"invalid file type '%v'",
			*outputType,
		)
	}
	fmt.Println(outputFile)
}

func NewFileName(outputPath, name, ext string) string {
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
}
