package main

type RunCmd struct {
	Policy string `arg:"" name:"policy" help:"Policy to run." type:"string"`
}

func (c *RunCmd) Run(g *Globals) error {
	err := g.App.Run(c.Policy)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("There was an error running the OPA runtime.")
	}

	<-g.App.Context.Done()

	return nil
}
