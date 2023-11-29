package main

import "github.com/opcr-io/policy/pkg/errors"

type SaveCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to save."`
	File   string `name:"file" short:"f" help:"Output file path, '-' is accepted for stdout" default:"bundle.tar.gz"`
}

func (c *SaveCmd) Run(g *Globals) error {
	err := g.App.Save(c.Policy, c.File)
	if err != nil {
		return errors.SaveFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
