package main

type ImagesCmd struct {
	Remote bool   `name:"remote" help:"List policies from a remote registry."`
	Sever  string `name:"server" short:"s" help:"Registry server to connect to" default:"opcr.io"`
}

func (c *ImagesCmd) Run(g *Globals) error {
	if c.Remote {
		err := g.App.ImagesRemote(c.Sever)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to list remote policies.")
		}
	} else {
		err := g.App.Images()
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to list local policies.")
		}
	}

	<-g.App.Context.Done()

	return nil
}
