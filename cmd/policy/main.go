package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/alecthomas/kong"
	"github.com/aserto-dev/logger"
	"github.com/opcr-io/policy/pkg/app"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/pkg/errors"
)

const (
	appName string = "policy"
)

var tmpConfig *config.Config

func ConfigExpander() kong.Resolver {
	var f kong.ResolverFunc = func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		resolveTmpConfig(context)
		t, err := template.New("value").Parse(flag.Default)
		if err != nil {
			return nil, err
		}

		t = t.Funcs(sprig.TxtFuncMap())
		buf := &bytes.Buffer{}
		err = t.Execute(buf, tmpConfig)
		if err != nil {
			return nil, err
		}
		expanded := buf.String()
		if expanded != flag.Default {
			flag.Default = expanded
			return expanded, nil
		}

		return nil, nil
	}

	return f
}

func resolveTmpConfig(context *kong.Context) {
	if tmpConfig != nil {
		return
	}

	allFlags := context.Flags()
	var configFlag *kong.Flag
	for _, f := range allFlags {
		if f.Name == "config" {
			configFlag = f
		}
	}

	if configFlag == nil {
		return
	}

	configPath := context.FlagValue(configFlag).(string)

	cfgLogger, err := config.NewLoggerConfig(config.Path(configPath), nil)
	if err != nil {
		panic(err)
	}

	log, err := logger.NewLogger(io.Discard, io.Discard, cfgLogger)
	if err != nil {
		panic(err)
	}

	tmpConfig, err = config.NewConfig(
		config.Path(configPath),
		log,
		nil)
	if err != nil {
		panic(err)
	}
}

type Globals struct {
	Debug     bool
	Config    string
	Verbosity int
	Insecure  bool
	App       *app.PolicyApp
}

var g *Globals

var PolicyCLI struct {
	Config    string `short:"c" type:"path" help:"Path to the policy CLI config file." default:"${userHome}/.config/policy/config.yaml"`
	Debug     bool   `help:"Enable debug mode."`
	Verbosity int    `short:"v" type:"counter" help:"Use to increase output verbosity."`
	Insecure  bool   `short:"k" help:"Do not verify TLS connections."`

	Build     BuildCmd     `cmd:"" help:"Build policies."`
	Images    ImagesCmd    `cmd:"" help:"List policy images."`
	Push      PushCmd      `cmd:"" help:"Push policies to a registry."`
	Pull      PullCmd      `cmd:"" help:"Pull policies from a registry."`
	Login     LoginCmd     `cmd:"" help:"Login to a registry."`
	Logout    LogoutCmd    `cmd:"" help:"Logout from a registry."`
	Save      SaveCmd      `cmd:"" help:"Save a policy to a local bundle tarball."`
	Tag       TagCmd       `cmd:"" help:"Create a new tag for an existing policy."`
	Rm        RmCmd        `cmd:"" help:"Removes a policy from the local registry."`
	Inspect   InspectCmd   `cmd:"" help:"Displays information about a policy."`
	Repl      ReplCmd      `cmd:"" help:"Sets you up with a shell for running queries using an OPA instance with a policy loaded."`
	Templates TemplatesCmd `cmd:"" help:"List and apply templates"`
	Version   VersionCmd   `cmd:"" help:"Prints version information."`
}

func (g *Globals) setup() func() {
	configFile := g.Config

	policyAPP, cleanup, err := app.BuildPolicyApp(
		os.Stderr,
		os.Stderr,
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
			c.Insecure = g.Insecure
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, `Application setup failed: %+v.
This might be a bug. Please open an issue here: https://github.com/opcr-io/policy\n`,
			err)
	}

	g.App = policyAPP
	return cleanup
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrap(err, "failed to determine user home directory"))
	}

	ctx := kong.Parse(
		&PolicyCLI,
		kong.Name(appName),
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
		kong.Resolvers(ConfigExpander()), kong.Vars{"userHome": home},
	)

	g = &Globals{
		Debug:     PolicyCLI.Debug,
		Config:    PolicyCLI.Config,
		Verbosity: PolicyCLI.Verbosity,
		Insecure:  PolicyCLI.Insecure,
	}

	cleanup := g.setup()

	if err := ctx.Run(g); err != nil {
		g.App.UI.Problem().Msg(err.Error())
		cleanup()
		os.Exit(1)
	}

	cleanup()
}
