package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/aserto-dev/policy-cli/pkg/version"
)

func main() {
	cliApp := &cli.App{
		Name:    "policy",
		Usage:   "policy",
		Version: version.GetInfo().String(),
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   filepath.Join(os.ExpandEnv("$HOME"), ".config", "aserto", "policy-cli", "config.yaml"),
				Usage:   "path of the configuration file",
			},
		},
		Commands: []*cli.Command{
			buildCmd,
			tagCmd,
			pushCmd,
			pullCmd,
			loginCmd,
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
