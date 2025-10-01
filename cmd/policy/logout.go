package main

import "github.com/opcr-io/policy/pkg/errors"

type LogoutCmd struct {
	Server string `name:"server" short:"s" help:"Server to logout from." default:"{{ .DefaultDomain }}"`
}

func (c *LogoutCmd) Run(g *Globals) error {
	g.App.UI.Normal().
		WithStringValue("server", c.Server).
		Msg("Logging out.")

	if err := g.App.RemoveServerCreds(c.Server); err != nil {
		return errors.ErrLogoutFailed.WithError(err)
	}

	g.App.UI.Normal().Msg("OK.")

	<-g.App.Context.Done()

	return nil
}
