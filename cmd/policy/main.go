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

var g *Globals

var PolicyCLI struct {
	Config    string `short:"c" type:"path" help:"Path to the policy CLI config file." default:"$HOME/.config/policy/config.yaml"`
	Debug     bool   `help:"Enable debug mode."`
	Verbosity int    `short:"v" type:"counter" help:"Use to increase output verbosity."`

	Build   BuildCmd   `cmd:"" help:"Build policies."`
	Images  ImagesCmd  `cmd:"" help:"List policy images."`
	Push    PushCmd    `cmd:"" help:"Push policies to a registry."`
	Pull    PullCmd    `cmd:"" help:"Pull policies from a registry."`
	Login   LoginCmd   `cmd:"" help:"Login to a registry."`
	Save    SaveCmd    `cmd:"" help:"Save a policy to a local bundle tarball."`
	Tag     TagCmd     `cmd:"" help:"Create a new tag for an existing policy."`
	Rm      RmCmd      `cmd:"" help:"Removes a policy from the local registry."`
	Run     RunCmd     `cmd:"" help:"Sets you up with a shell for running queries using an OPA instance with a policy loaded."`
	Version VersionCmd `cmd:"" help:"Prints version information."`
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
	ctx := kong.Parse(&PolicyCLI, kong.Resolvers(
		EnvExpander()))

	g = &Globals{
		Debug:     PolicyCLI.Debug,
		Config:    PolicyCLI.Config,
		Verbosity: PolicyCLI.Verbosity,
	}
	cleanup := g.setup()
	defer cleanup()

	err := ctx.Run(g)

	ctx.FatalIfErrorf(err)
}
