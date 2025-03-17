package main

import (
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/docker/cli/cli/config/types"

	perr "github.com/opcr-io/policy/pkg/errors"
	"golang.org/x/term"
)

type LoginCmd struct {
	Server        string `name:"server" short:"s" help:"Server to connect to." default:"{{ .DefaultDomain }}"`
	Username      string `name:"username" required:"" short:"u" help:"Username for logging into the server."`
	Password      string `name:"password" short:"p" help:"Password for logging into the server."`
	PasswordStdin bool   `name:"password-stdin" help:"Take the password from stdin"`
	DefaultDomain bool   `name:"default-domain" short:"d" help:"Do not ask for setting default domain"`
}

func (c *LoginCmd) Run(g *Globals) error {
	if c.Password != "" {
		g.App.UI.Exclamation().Msg("Using --password via the CLI is insecure. Use --password-stdin.")

		if c.PasswordStdin {
			return perr.LoginFailed.WithMessage("--password and --password-stdin are mutually exclusive")
		}
	}

	if c.PasswordStdin {
		if c.Username == "" {
			return perr.LoginFailed.WithMessage("Must provide --username with --password-stdin")
		}

		contents, err := io.ReadAll(g.App.UI.Input())
		if err != nil {
			return perr.LoginFailed.WithError(err)
		}

		c.Password = strings.TrimSuffix(string(contents), "\n")
		c.Password = strings.TrimSuffix(c.Password, "\r")
	}

	if c.Server == "" {
		return perr.LoginFailed.WithMessage("Must provide --server")
	}

	password := c.Password
	if c.Password == "" {
		g.App.UI.Normal().NoNewline().Msg("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin)) // nolint:unconvert // needed for windows
		if err != nil {
			return perr.LoginFailed.WithError(err)
		}

		password = string(bytePassword)
	}

	g.App.UI.Normal().
		WithStringValue("server", c.Server).
		WithStringValue("user", c.Username).
		Msg("Logging in.")
	if err := g.App.Ping(c.Server, c.Username, password); err != nil {
		return perr.LoginFailed.WithError(err)
	}

	var setDefault bool
	stat, err := os.Stdin.Stat()
	if err != nil {
		return perr.LoginFailed.WithError(err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		setDefault = c.DefaultDomain
	} else {
		setDefault = checkDefault(g, c)
	}

	err = g.App.SaveServerCreds(&types.AuthConfig{
		ServerAddress: c.Server,
		Auth:          "basic",
		Username:      c.Username,
		Password:      password,
	})
	if err != nil {
		return perr.LoginFailed.WithError(err)
	}

	if setDefault {
		g.App.Configuration.DefaultDomain = c.Server
	}

	err = g.App.Configuration.SaveDefaultDomain()
	if err != nil {
		return perr.LoginFailed.WithError(err)
	}

	g.App.UI.Normal().Msg("OK.")

	<-g.App.Context.Done()
	return nil
}

func checkDefault(g *Globals, c *LoginCmd) bool {
	setDefault := c.DefaultDomain
	if c.Server != g.App.Configuration.DefaultDomain && !c.DefaultDomain {
		g.App.UI.Normal().WithAskBoolMap("Do you want to set this server as your default domain?[yes/no]", &setDefault, map[string]bool{
			"yes": true,
			"no":  false,
			"y":   true,
			"n":   false,
			"Y":   true,
			"N":   false,
		}).Do()
	} else {
		// already on default server
		return true
	}

	return setDefault
}
