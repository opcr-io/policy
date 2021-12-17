package main

type RmCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to remove from the local registry."`
	Remote   bool     `name:"remote" short:"r" help:"Remove the policy from the remote server."`
	All      bool     `name:"all" short:"a" help:"When remote is set, remove all tags and the policy reference."`
	Force    bool     `name:"force" short:"f" help:"Don't ask for confirmation."`
}

func (c *RmCmd) Run(g *Globals) error {
	for _, policyRef := range c.Policies {
		if c.Remote {
			err := g.App.RmRemote(policyRef, c.All, c.Force)
			if err != nil {
				g.App.UI.Problem().WithErr(err).Msg("Failed to remove policy.")
				return err
			}
		} else {
			err := g.App.Rm(policyRef, c.Force)
			if err != nil {
				g.App.UI.Problem().WithErr(err).Msg("Failed to remove policy.")
				return err
			}
		}
	}

	<-g.App.Context.Done()

	return nil
}
