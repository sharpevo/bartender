package options

import (
	"flag"
)

type TransferOptions struct {
	InputPath string
}

func AttachTransferOptions(cmd *flag.FlagSet) *TransferOptions {
	options := &TransferOptions{}
	cmd.StringVar(
		&options.InputPath,
		"sourcepath",
		"",
		"local directory to transfer",
	)
	return options
}
