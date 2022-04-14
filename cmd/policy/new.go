package main

type NewCmd struct {
	TemplateName string `arg:"" name:"template" required:"" help:"the name of the template to use" default:""`
	List         bool   `name:"list" short:"l" help:"List all available policy templates."`
	Output       string `name:"output" short:"o" help:"output directory" default:"."`
	Overwrite    bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
}

func (c *NewCmd) Run(g *Globals) error {
	// err := g.App.New(c.TemplateName, c.Output, c.List, c.Overwrite)
	// if err != nil {
	// 	return errors.Wrap(err, "Failed create policy")
	// }

	return nil
}
