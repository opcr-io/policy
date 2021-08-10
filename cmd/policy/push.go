package main

type PushCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to push."`
}

func (c *PushCmd) Run(g *Globals) error {
	cleanup := g.setup()
	defer cleanup()

	for _, policyRef := range c.Policies {
		err := g.App.Push(policyRef)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msg("Failed to push policy.")
		}
	}

	<-g.App.Context.Done()

	return nil
}
