package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/opcr-io/policy/pkg/cmd"
	"github.com/pkg/errors"
)

const (
	rcOK  int = 0
	rcErr int = 1
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	os.Exit(run())
}

func exitErr(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return rcErr
}

func run() int {
	cli := cmd.CLI{}

	home, err := os.UserHomeDir()
	if err != nil {
		return exitErr(errors.Wrap(err, "failed to determine user home directory"))
	}

	kongCtx := kong.Parse(
		&cli,
		kong.Name(cmd.AppName),
		kong.UsageOnError(),
		kong.Exit(func(x int) { os.Exit(0) }),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Summary:             false,
			Compact:             true,
			Tree:                false,
			FlagsLast:           true,
			Indenter:            kong.SpaceIndenter,
			NoExpandSubcommands: true,
		}),
		kong.Resolvers(cmd.ConfigExpander()), kong.Vars{"userHome": home},
	)

	g := &cmd.Globals{
		Debug:     cli.Debug,
		Config:    cli.Config,
		Verbosity: cli.Verbosity,
		Insecure:  cli.Insecure,
		Plaintext: cli.Plaintext,
	}

	cleanup := g.Setup()
	defer cleanup()

	if err := kongCtx.Run(g); err != nil {
		return exitErr(err)
	}

	return rcOK
}
