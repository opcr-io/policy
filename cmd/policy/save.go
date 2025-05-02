package main

import perr "github.com/opcr-io/policy/pkg/errors"

type SaveCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to save."`
	File   string `name:"file" short:"f" help:"Output file path, '-' is accepted for stdout" default:"bundle.tar.gz"`
}

func (c *SaveCmd) Run(g *Globals) error {
	if err := g.App.Save(c.Policy, c.File); err != nil {
		return perr.ErrPolicySaveFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
