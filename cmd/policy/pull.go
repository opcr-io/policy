package main

import (
	"errors"
)

type PullCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to pull from the remote registry."`
}

func (c *PullCmd) Run(g *Globals) error {
	var errs error
	for _, policyRef := range c.Policies {
		err := g.App.Pull(policyRef)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msgf("Failed to pull policy: %s", policyRef)
			errs = err
		}
	}

	<-g.App.Context.Done()

	if errs != nil {
		return errors.New("failed to pull one or more policies")
	}

	return nil
}
