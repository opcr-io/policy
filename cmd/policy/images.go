package main

type ImagesCmd struct {
	Remote    bool   `name:"remote" short:"r" help:"List policies from a remote registry."`
	Server    string `name:"server" short:"s" help:"Registry server to connect to" default:"opcr.io"`
	ShowEmpty bool   `name:"show-empty" short:"e" help:"Show policies with no images."`
}

func (c *ImagesCmd) Run(g *Globals) error {
	if c.Remote {
		err := g.App.ImagesRemote(c.Server, c.ShowEmpty)
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
