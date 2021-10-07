package main

type LogoutCmd struct {
	Server string `name:"server" short:"s" help:"Server to logout from." default:"opcr.io"`
}

func (c *LogoutCmd) Run(g *Globals) error {
	g.App.UI.Normal().
		WithStringValue("server", c.Server).
		Msg("Logging out.")

	err := g.App.RemoveServerCreds(c.Server)
	if err != nil {
		return err
	}

	g.App.UI.Normal().Msg("OK.")

	<-g.App.Context.Done()
	return nil
}
