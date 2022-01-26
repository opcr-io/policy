package main

import "github.com/pkg/errors"

type InitCmd struct {
	RootPath     string `arg:"" name:"path" required:"" help:"project root path"`
	User         string `name:"user" required:"" help:"user name"`
	Server       string `name:"server" required:"" help:"registry service name"`
	Repo         string `name:"repo" required:"" help:"repo name like org/repo"`
	SCC          string `name:"scc" required:"" help:"source code provider [github]" default:"github"`
	ActionSecret string `name:"secret" required:"" help:"Github action secret name"`
	Overwrite    bool   `name:"overwrite" help:"overwrite existing default:false"`
}

func (c *InitCmd) Run(g *Globals) error {
	err := g.App.Init(c.RootPath, c.User, c.Server, c.Repo, c.SCC, c.ActionSecret, c.Overwrite)
	if err != nil {
		return errors.Wrap(err, "Init failed.")
	}

	<-g.App.Context.Done()

	return nil
}
