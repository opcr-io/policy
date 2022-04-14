package main

import "github.com/pkg/errors"

type TemplatesCmd struct {
	Apply ApplyCmd `cmd:"" name:"apply" help:"Create or update a policy or related artifacts from a template."`
	List  ListCmd  `cmd:"" name:"list" help:"List all available templates."`
}

type ApplyCmd struct {
	Output    string `arg:"" name:"path" required:"" help:"output directory (defaults to current directory)" default:"."`
	Overwrite bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
}

type ListCmd struct {
}

func (c *ApplyCmd) Run(g *Globals) error {
	return nil
}

func (c *ListCmd) Run(g *Globals) error {
	err := g.App.TemplatesList()
	if err != nil {
		return errors.Wrap(err, "Failed list templates")
	}

	return nil
}
