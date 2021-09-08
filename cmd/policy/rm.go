package main

type RmCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to remove from the local registry."`
}

func (c *RmCmd) Run(g *Globals) error {
	for _, policyRef := range c.Policies {
		err := g.App.Rm(policyRef)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to remove policy.")
		}
	}

	<-g.App.Context.Done()

	return nil
}
