package main

import "github.com/pkg/errors"

type InitCmd struct {
	RootPath     string `arg:"" name:"path" required:"" help:"project root path"`
	User         string `name:"user" short:"u" required:"" help:"user name"`
	Server       string `name:"server" short:"s" required:"" help:"registry service name"`
	Repo         string `name:"repo" short:"r" required:"" help:"repository (org/repo)"`
	SCC          string `name:"scc" required:"" help:"source code provider" enum:"github" default:"github"`
	ActionSecret string `name:"secret" required:"" help:"Github action secret name" default:"GITHUB_TOKEN"`
	Overwrite    bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
}

func (c *InitCmd) Run(g *Globals) error {
	err := g.App.Init(c.RootPath, c.User, c.Server, c.Repo, c.SCC, c.ActionSecret, c.Overwrite)
	if err != nil {
		return errors.Wrap(err, "Init failed.")
	}

	<-g.App.Context.Done()

	return nil
}
