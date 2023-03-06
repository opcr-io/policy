package main

import "github.com/pkg/errors"

type RmCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to remove from the local registry."`
	Remote   bool     `name:"remote" short:"r" help:"Remove the policy from the remote server."`
	All      bool     `name:"all" short:"a" help:"When remote is set, remove all tags and the policy reference."`
	Force    bool     `name:"force" short:"f" help:"Don't ask for confirmation."`
}

func (c *RmCmd) Run(g *Globals) error {
	var errs error
	for _, policyRef := range c.Policies {
		if c.Remote {
			g.App.UI.Exclamation().Msg("This command is deprecated and it will be removed in a future version of the policy CLI.")
			err := g.App.RmRemote(policyRef, c.All, c.Force)
			if err != nil {
				g.App.UI.Problem().WithErr(err).Msgf("Failed to remove policy: %s", policyRef)
				errs = err
			}
		} else {
			err := g.App.Rm(policyRef, c.Force)
			if err != nil {
				g.App.UI.Problem().WithErr(err).Msgf("Failed to remove policy: %s", policyRef)
				errs = err
			}
		}
	}

	<-g.App.Context.Done()
	if errs != nil {
		return errors.Wrap(errs, "failed to remove one or more policies")
	}

	return nil
}
