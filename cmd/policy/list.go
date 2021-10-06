package main

type ImagesCmd struct {
	Remote bool   `name:"remote" help:"List policies from a remote registry."`
	Sever  string `name:"server" short:"s" help:"Registry server to connect to" default:"opcr.io"`
}

func (c *ImagesCmd) Run(g *Globals) error {
	if c.Remote {
		err := g.App.ListRemote(c.Sever)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to list local policies.")
		}

		return nil
	} else {
		err := g.App.List()
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to list local policies.")
		}
	}
	<-g.App.Context.Done()

	return nil
}
