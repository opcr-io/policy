package main

import (
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/pkg/errors"
	"golang.org/x/term"
)

type LoginCmd struct {
	Server        string `name:"server" short:"s" help:"Server to connect to." default:"{{ .DefaultDomain }}"`
	Username      string `name:"username" short:"u" help:"Username for logging into the server."`
	Password      string `name:"password" short:"p" help:"Password for logging into the server."`
	PasswordStdin bool   `name:"password-stdin" help:"Take the password from stdin"`
}

func (c *LoginCmd) Run(g *Globals) error {
	if c.Password != "" {
		g.App.UI.Exclamation().Msg("Using --password via the CLI is insecure. Use --password-stdin.")

		if c.PasswordStdin {
			return errors.New("--password and --password-stdin are mutually exclusive")
		}
	}

	if c.PasswordStdin {
		if c.Username == "" {
			return errors.New("Must provide --username with --password-stdin")
		}

		contents, err := ioutil.ReadAll(g.App.UI.Input())
		if err != nil {
			return err
		}

		c.Password = strings.TrimSuffix(string(contents), "\n")
		c.Password = strings.TrimSuffix(c.Password, "\r")
	}

	password := c.Password
	if c.Password == "" {
		g.App.UI.Normal().NoNewline().Msg("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin)) // nolint:unconvert // needed for windows
		if err != nil {
			return errors.Wrap(err, "failed to read password from stdin")
		}

		password = string(bytePassword)
	}

	g.App.UI.Normal().
		WithStringValue("server", c.Server).
		WithStringValue("user", c.Username).
		Msg("Logging in.")
	err := g.App.Ping(c.Server, c.Username, password)
	if err != nil {
		return err
	}

	err = g.App.SaveServerCreds(c.Server, config.ServerCredentials{
		Type:     "basic",
		Username: c.Username,
		Password: password,
	})
	if err != nil {
		return err
	}

	g.App.UI.Normal().Msg("OK.")

	<-g.App.Context.Done()
	return nil
}
