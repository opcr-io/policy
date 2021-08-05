package main

import (
	"os"

	"github.com/aserto-dev/policy-cli/pkg/app"
	"github.com/aserto-dev/policy-cli/pkg/cc/config"
	"github.com/urfave/cli/v2"
)

var tagCmd = &cli.Command{
	Name:  "tag",
	Usage: "tag",
	Action: func(c *cli.Context) error {

		configFile := c.String("config")

		app, cleanup, err := app.BuildPolicyCLI(
			os.Stdout,
			config.Path(configFile),
			func(*config.Config) {})

		defer func() {
			if cleanup != nil {
				cleanup()
			}
		}()
		if err != nil {
			return err
		}

		<-app.Context.Done()

		return nil
	},
}
