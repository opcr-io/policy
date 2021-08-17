package main

type SaveCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to save."`
	File   string `name:"file" short:"f" help:"Output file path" default:"bundle.tar.gz"`
}

func (c *SaveCmd) Run(g *Globals) error {
	err := g.App.Save(c.Policy, c.File)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Failed to save local bundle tarball.")
	}

	<-g.App.Context.Done()

	return nil
}
