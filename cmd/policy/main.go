package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/policy/pkg/app"
	"github.com/aserto-dev/policy/pkg/cc/config"
)

func EnvExpander() kong.Resolver {
	var f kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		expanded := os.ExpandEnv(flag.Default)
		if expanded != flag.Default {
			flag.Default = expanded
			return expanded, nil
		}

		return nil, nil
	}

	return f
}

type Globals struct {
	Debug     bool
	Config    string
	Verbosity int
	App       *app.PolicyApp
}

var PolicyCLI struct {
	Config    string `short:"c" type:"path" help:"Path to the policy CLI config file." default:"$HOME/.config/aserto/policy/config.yaml"`
	Debug     bool   `help:"Enable debug mode."`
	Verbosity int    `short:"v" type:"counter" help:"Use to increase output verbosity."`

	Build BuildCmd `cmd:"" help:"Build policies."`
	List  ListCmd  `cmd:"" help:"List policies."`
	Push  PushCmd  `cmd:"" help:"Push policies to a registry."`
}

func (g *Globals) setup() func() {
	configFile := g.Config

	policyAPP, cleanup, err := app.BuildPolicyApp(
		os.Stdout,
		config.Path(configFile),
		func(c *config.Config) {
			switch g.Verbosity {
			case 0:
				c.Logging.LogLevel = "error"
			case 1:
				c.Logging.LogLevel = "info"
			case 2:
				c.Logging.LogLevel = "debug"
			default:
				c.Logging.LogLevel = "trace"
			}
		})

	if err != nil {
		fmt.Printf(`Application setup failed: %+v.
This might be a bug. Please open an issue here: https://github.com/aserto-dev/policy\n`,
			err)
	}

	g.App = policyAPP
	return cleanup
}

func main() {
	ctx := kong.Parse(&PolicyCLI, kong.Resolvers(EnvExpander()))
	err := ctx.Run(&Globals{
		Debug:     PolicyCLI.Debug,
		Config:    PolicyCLI.Config,
		Verbosity: PolicyCLI.Verbosity,
	})

	ctx.FatalIfErrorf(err)
}
