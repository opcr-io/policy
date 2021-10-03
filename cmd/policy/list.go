package main

type ImagesCmd struct {
}

func (c *ImagesCmd) Run(g *Globals) error {
	err := g.App.List()
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Failed to list local policies.")
	}

	<-g.App.Context.Done()

	return nil
}
