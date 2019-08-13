package options

import (
	"flag"
)

type Options struct {
	InputPath string
}

func AttachOptions(cmd *flag.FlagSet) *Options {
	options := &Options{}
	cmd.StringVar(
		&options.InputPath,
		"sourcepath",
		"",
		"local directory to transfer",
	)
	return options
}
