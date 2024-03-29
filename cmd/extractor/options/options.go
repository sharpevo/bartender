package options

import (
	"flag"
)

type Options struct {
	InputPath      string
	SheetIndex     int
	RowStartsAt    int
	RowEndsAt      int
	ColumnIndices  string
	OutputType     string
	OutputPath     string
	ExtractPattern string
}

func AttachOptions(cmd *flag.FlagSet) *Options {
	options := &Options{}
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
	cmd.StringVar(
		&options.ExtractPattern,
		"extractpattern",
		"",
		"filename pattern to extract",
	)
	return options
}
