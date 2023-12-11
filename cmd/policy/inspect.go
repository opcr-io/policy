package main

import "github.com/opcr-io/policy/pkg/errors"

type InspectCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to inspect."`
}

func (c *InspectCmd) Run(g *Globals) error {
	err := g.App.Inspect(c.Policy)
	if err != nil {
		return errors.InspectFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
