package main

import "github.com/pkg/errors"

type ReplCmd struct {
	Policy    string `arg:"" name:"policy" help:"Policy to run." type:"string"`
	MaxErrors int    `name:"max-errors" short:"m" help:"Set the number of errors to allow before compilation fails early." default:"10"`
}

func (c *ReplCmd) Run(g *Globals) error {
	if err := g.App.Repl(c.Policy, c.MaxErrors); err != nil {
		return errors.Wrap(err, "there was an error running the OPA runtime")
	}

	<-g.App.Context.Done()

	return nil
}
