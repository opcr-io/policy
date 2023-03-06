package main

import "github.com/pkg/errors"

type RmCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to remove from the local registry."`
	All      bool     `name:"all" short:"a" help:"When remote is set, remove all tags and the policy reference."`
	Force    bool     `name:"force" short:"f" help:"Don't ask for confirmation."`
}

func (c *RmCmd) Run(g *Globals) error {
	var errs error
	for _, policyRef := range c.Policies {
		err := g.App.Rm(policyRef, c.Force)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msgf("Failed to remove policy: %s", policyRef)
			errs = err
		}
	}

	<-g.App.Context.Done()
	if errs != nil {
		return errors.Wrap(errs, "failed to remove one or more policies")
	}

	return nil
}
