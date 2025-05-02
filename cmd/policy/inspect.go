package main

import perr "github.com/opcr-io/policy/pkg/errors"

type InspectCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to inspect."`
}

func (c *InspectCmd) Run(g *Globals) error {
	if err := g.App.Inspect(c.Policy); err != nil {
		return perr.ErrPolicyInspectFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
