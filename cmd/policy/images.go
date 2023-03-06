package main

import (
	"strings"

	"github.com/pkg/errors"
)

type ImagesCmd struct {
	Remote    bool   `name:"remote" short:"r" help:"List policies from a remote registry."`
	Server    string `name:"server" short:"s" help:"Registry server to connect to." default:"{{ .DefaultDomain }}"`
	ShowEmpty bool   `name:"show-empty" short:"e" help:"Show policies with no images."`
	Org       string `name:"organization" short:"o" help:"Show images for an organization."`
}

func (c *ImagesCmd) Run(g *Globals) error {
	if c.Remote {
		g.App.UI.Exclamation().Msg("This command is deprecated and it will be removed in a future version of the policy CLI.")
		if c.Org == "" {
			return errors.New("organization parameter is required, please provide it using -o/--organization")
		}
		err := g.App.ImagesRemote(c.Server, strings.TrimSpace(c.Org), c.ShowEmpty)
		if err != nil {
			return errors.Wrap(err, "failed to list remote policies")
		}
	} else {
		err := g.App.Images()
		if err != nil {
			return errors.Wrap(err, "failed to list local policies")
		}
	}

	<-g.App.Context.Done()

	return nil
}
