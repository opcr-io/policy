package main

import "github.com/aserto-dev/policy/pkg/cc/config"

type LoginCmd struct {
	Server   string `name:"server" short:"s" help:"Server to connect to."`
	Username string `name:"username" short:"u" help:"Username for logging into the server."`
	Password string `name:"password" short:"p" help:"Password for logging into the server."`
}

func (c *LoginCmd) Run(g *Globals) error {
	err := g.App.SaveServerCreds(c.Server, config.ServerCredentials{
		Type:     "basic",
		Username: c.Username,
		Password: c.Password,
	})

	if err != nil {
		return err
	}

	<-g.App.Context.Done()
	return nil
}
