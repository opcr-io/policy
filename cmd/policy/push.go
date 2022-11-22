package main

import "github.com/pkg/errors"

type PushCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to push."`
}

func (c *PushCmd) Run(g *Globals) error {
	var errs error
	for _, policyRef := range c.Policies {
		err := g.App.Push(policyRef)
		if err != nil {
			g.App.UI.Problem().WithErr(err).Msgf("Failed to push policy: %s", policyRef)
			errs = err
		}
	}

	<-g.App.Context.Done()

	if errs != nil {
		return errors.Wrap(errs, "failed to push one or more policies")
	}

	return nil
}
