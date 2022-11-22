package main

import "github.com/pkg/errors"

type TagCmd struct {
	Policy string `arg:"" name:"policy" help:"Source policy name." type:"string"`
	Tag    string `arg:"" name:"tag" help:"Name and optionally a tag in the 'name:tag' format"`
}

func (c *TagCmd) Run(g *Globals) error {
	err := g.App.Tag(c.Policy, c.Tag)
	if err != nil {
		return errors.Wrap(err, "tagging failed")
	}

	<-g.App.Context.Done()

	return nil
}
