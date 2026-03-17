package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/opcr-io/policy/pkg/cmd"
	"github.com/pkg/errors"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "failed to determine user home directory"))
	}

	ctx := kong.Parse(
		&cmd.CLI,
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
		Debug:     cmd.CLI.Debug,
		Config:    cmd.CLI.Config,
		Verbosity: cmd.CLI.Verbosity,
		Insecure:  cmd.CLI.Insecure,
		Plaintext: cmd.CLI.Plaintext,
	}

	cleanup := g.Setup()

	if err := ctx.Run(g); err != nil {
		g.App.UI.Problem().Msg(err.Error())
		cleanup()
		os.Exit(1)
	}

	cleanup()
}
