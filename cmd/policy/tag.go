package main

import "github.com/opcr-io/policy/pkg/errors"

type TagCmd struct {
	Policy string `arg:"" name:"policy" help:"Source policy name." type:"string"`
	Tag    string `arg:"" name:"tag" help:"Name and optionally a tag in the 'name:tag' format"`
}

func (c *TagCmd) Run(g *Globals) error {
	err := g.App.Tag(c.Policy, c.Tag)
	if err != nil {
		return errors.ErrTagFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
