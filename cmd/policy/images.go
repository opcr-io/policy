package main

import (
	"github.com/opcr-io/policy/pkg/errors"
)

type ImagesCmd struct {
	Server    string `name:"server" short:"s" help:"Registry server to connect to." default:"{{ .DefaultDomain }}"`
	ShowEmpty bool   `name:"show-empty" short:"e" help:"Show policies with no images."`
	Org       string `name:"organization" short:"o" help:"Show images for an organization."`
}

func (c *ImagesCmd) Run(g *Globals) error {

	err := g.App.Images()
	if err != nil {
		return errors.ImagesFailed.WithError(err)
	}

	<-g.App.Context.Done()

	return nil
}
