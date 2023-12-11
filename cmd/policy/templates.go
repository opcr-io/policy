package main

import "github.com/opcr-io/policy/pkg/errors"

type TemplatesCmd struct {
	Apply ApplyCmd `cmd:"" name:"apply" help:"Create or update a policy or related artifacts from a template."`
	List  ListCmd  `cmd:"" name:"list" help:"List all available templates."`
}

type ApplyCmd struct {
	Template  string `arg:"" name:"template" required:"true" help:"name of the template to apply"`
	Output    string `name:"output" short:"o" help:"output directory (defaults to current directory)" default:"."`
	Overwrite bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
}

type ListCmd struct {
}

func (c *ApplyCmd) Run(g *Globals) error {
	err := g.App.TemplateApply(c.Template, c.Output, c.Overwrite)
	if err != nil {
		return errors.TemplateFailed.WithError(err)
	}

	<-g.App.Context.Done()
	return nil
}

func (c *ListCmd) Run(g *Globals) error {
	err := g.App.TemplatesList()
	if err != nil {
		return errors.TemplateFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
