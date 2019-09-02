package options

import (
	"flag"
	"fmt"
	"time"
)

func Debug(i interface{}) string {
	return fmt.Sprintf("%+v", i)
}

type ServerOptions struct {
	Enabled   bool
	HostKey   string
	UserName  string
	Password  string
	Directory string
}

func AttachServerOptions(cmd *flag.FlagSet) *ServerOptions {
	options := &ServerOptions{}
	cmd.BoolVar(
		&options.Enabled,
		"transfer",
		false,
		"enable output file transfer",
	)
	cmd.StringVar(
		&options.HostKey,
		"hostkey",
		"101.201.180.67 ecdsa-sha2-nistp256 xxx",
		"lines in ./ssh/known_host",
	)
	cmd.StringVar(
		&options.UserName,
		"username",
		"root",
		"user name to use when connecting to remote server",
	)
	cmd.StringVar(
		&options.Password,
		"password",
		"",
		"password to use when connecting to remote server",
	)
	cmd.StringVar(
		&options.Directory,
		"remotepath",
		"",
		"path of output directory",
	)
	return options
}

type WatchOptions struct {
	InputPath       string
	Enabled         bool
	Interval        time.Duration
	FileNamePattern string
}

func AttachWatchOptions(cmd *flag.FlagSet) *WatchOptions {
	options := &WatchOptions{}
	cmd.BoolVar(
		&options.Enabled,
		"watch",
		false,
		"enable to watch the directory",
	)
	cmd.DurationVar(
		&options.Interval,
		"interval",
		10*time.Second,
		"interval of walking through the folders, not for files",
	)
	cmd.StringVar(
		&options.InputPath,
		"watchpath",
		"",
		"path of watched direcotry",
	)
	cmd.StringVar(
		&options.FileNamePattern,
		"namepattern",
		"",
		"filename pattern",
	)
	return options
}

type LogOptions struct {
	Level string
}

func AttachLogOptions(cmd *flag.FlagSet) *LogOptions {
	options := &LogOptions{}
	cmd.StringVar(
		&options.Level,
		"loglevel",
		"info",
		"log level",
	)
	return options
}
