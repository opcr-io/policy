package main

type BuildCmd struct {
	Tag  []string `name:"tag" short:"t" help:"Name and optionally a tag in the 'name:tag' format"`
	Path string   `arg:"" name:"path" help:"Path to the policy sources." type:"path"`
}

func (c *BuildCmd) Run(g *Globals) error {
	cleanup := g.setup()
	defer cleanup()

	err := g.App.Build(c.Tag, c.Path)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Build failed.")
	}

	<-g.App.Context.Done()

	return nil
}
