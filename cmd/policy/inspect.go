package main

import "github.com/pkg/errors"

type InspectCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to inspect."`
}

func (c *InspectCmd) Run(g *Globals) error {
	err := g.App.Inspect(c.Policy)
	if err != nil {
		return errors.Wrap(err, "failed to inspect policy")
	}

	<-g.App.Context.Done()

	return nil
}
