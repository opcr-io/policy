package main

import perr "github.com/opcr-io/policy/pkg/errors"

type TagCmd struct {
	Policy string `arg:"" name:"policy" help:"Source policy name." type:"string"`
	Tag    string `arg:"" name:"tag" help:"Name and optionally a tag in the 'name:tag' format"`
}

func (c *TagCmd) Run(g *Globals) error {
	if err := g.App.Tag(c.Policy, c.Tag); err != nil {
		return perr.TagFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
