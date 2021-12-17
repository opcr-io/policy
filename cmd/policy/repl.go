package main

type ReplCmd struct {
	Policy    string `arg:"" name:"policy" help:"Policy to run." type:"string"`
	MaxErrors int    `name:"max-errors" short:"m" help:"Set the number of errors to allow before compilation fails early." default:"10"`
}

func (c *ReplCmd) Run(g *Globals) error {
	err := g.App.Repl(c.Policy, c.MaxErrors)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("There was an error running the OPA runtime.")
		return err
	}

	<-g.App.Context.Done()

	return nil
}
