package main

import "github.com/pkg/errors"

type RemoteCmd struct {
	SetPublic SetPublicCmd    `cmd:"" name:"set-public" help:"Make a policy public or private."`
	Images    RemoteImagesCmd `cmd:"" help:"Synonym for 'policy images --remote'."`
}

type SetPublicCmd struct {
	Server string `name:"server" short:"s" help:"Registry server to connect to" default:"{{ .DefaultDomain }}"`
	Policy string `arg:"" name:"policy" help:"Policy to publish."`
	Public string `arg:"" default:"true" help:"Set to 'false' to make the policy private. Default is 'true' and makes a policy world-readable."`
}

type RemoteImagesCmd struct {
	Server    string `name:"server" short:"s" help:"Registry server to connect to" default:"{{ .DefaultDomain }}"`
	ShowEmpty bool   `name:"show-empty" short:"e" help:"Show policies with no images."`
}

func (c *SetPublicCmd) Run(g *Globals) error {
	public := false
	if c.Public == "true" {
		public = true
	} else if c.Public != "false" {
		return errors.Errorf("Invalid value for --public: [%s]", c.Public)
	}

	err := g.App.SetVisibility(c.Server, c.Policy, public)
	if err != nil {
		g.App.UI.Problem().WithErr(err).Msg("Failed to set policy visibility.")
	}

	<-g.App.Context.Done()

	return nil
}

func (c *RemoteImagesCmd) Run(g *Globals) error {
	return (&ImagesCmd{
		Remote:    true,
		Server:    c.Server,
		ShowEmpty: c.ShowEmpty,
	}).Run(g)
}
