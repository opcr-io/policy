package main

import (
	"fmt"
	"syscall"

	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/pkg/errors"
	"golang.org/x/term"
)

type LoginCmd struct {
	Server   string `name:"server" short:"s" help:"Server to connect to." default:"opcr.io"`
	Username string `name:"username" short:"u" help:"Username for logging into the server."`
	Password string `name:"password" short:"p" help:"Password for logging into the server."`
}

func (c *LoginCmd) Run(g *Globals) error {
	password := c.Password
	if c.Password == "" {
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(syscall.Stdin)
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
