package main

type ListCmd struct {
}

func (c *ListCmd) Run(g *Globals) error {
	cleanup := g.setup()
	defer cleanup()

	err := g.App.List()
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Failed to list local policies.")
	}

	<-g.App.Context.Done()

	return nil
}
