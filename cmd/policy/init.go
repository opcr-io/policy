package main

import "github.com/pkg/errors"

type InitCmd struct {
	RootPath  string `arg:"" name:"path" required:"" help:"project root path (defaults to current directory)" default:"."`
	User      string `name:"user" short:"u" help:"user name"`
	Server    string `name:"server" short:"s" help:"registry service name"`
	Repo      string `name:"repo" short:"r" help:"repository (org/repo)"`
	TokenName string `name:"token" short:"t" help:"Github Actions secret token name"`
	SCC       string `name:"scc" help:"source code provider"`
	Overwrite bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
	NoSrc     bool   `name:"no-src" help:"do not write src directory" default:"false"`
}

func (c *InitCmd) Run(g *Globals) error {
	return errors.New("policy init has been deprecated please use 'policy templates apply' instead")
}
