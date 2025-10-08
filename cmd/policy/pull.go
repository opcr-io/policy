package main

import "github.com/pkg/errors"

type PullCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to pull from the remote registry."`
	Untar    bool     `name:"untar" help:"Untar the policy bundle after pulling."`
	UntarDir string   `name:"untardir" help:"Directory to extract the policy bundle." default:"." type:"path"`
}

func (c *PullCmd) Run(g *Globals) error {
	var errs error

	for _, policyRef := range c.Policies {
		err := g.App.Pull(policyRef, c.Untar, c.UntarDir)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msgf("Failed to pull policy: %s", policyRef)
			errs = err
		}
	}

	<-g.App.Context.Done()

	if errs != nil {
		return errors.Wrap(errs, "failed to pull one or more policies")
	}

	return nil
}
