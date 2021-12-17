package main

type InspectCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to inspect."`
}

func (c *InspectCmd) Run(g *Globals) error {
	err := g.App.Inspect(c.Policy)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Failed to inspect policy.")
		return err
	}

	<-g.App.Context.Done()

	return nil
}
