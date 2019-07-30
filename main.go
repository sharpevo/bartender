package main

import (
	"excel"
	//"flag"
	"fmt"
	"os"
)

func main() {
	//fileNamep := flag.String("file", "test.xlsx", "")
	//sheetIndexp := flag.Int("sheet", "0", "sheet index")
	//rowStartsAtp := flag.Int("from", "2", "top row index of the range")
	//rowEndsAtp := flag.Int("to", "-1", "bottom row index of the range")
	//columnIndices := flag.String("columns", "0,2,10", "colmuns to be extracted")

	if len(os.Args) < 2 {
		fmt.Println("invaild file name")
		return
	}
	data, err := excel.ExtractColumns(
		os.Args[1],
		1,
		2,
		-1,
		[]int{0, 1, 2, 3, 4, 9, 11},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	if excel.MakeFileCSV(
		"output.csv",
		data,
		excel.SEPARATOR_TAB,
	) != nil {
		fmt.Println(err)
		return
	}
}
