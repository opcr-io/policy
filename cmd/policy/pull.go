package main

type PullCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to pull from the remote registry."`
}

func (c *PullCmd) Run(g *Globals) error {
	for _, policyRef := range c.Policies {
		err := g.App.Pull(policyRef)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to pull policy.")
		}
	}

	<-g.App.Context.Done()

	return nil
}
