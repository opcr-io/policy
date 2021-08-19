package main

import "github.com/aserto-dev/policy/pkg/version"

type VersionCmd struct {
}

func (c *VersionCmd) Run(g *Globals) error {
	v := version.GetInfo()

	g.App.UI.Normal().
		WithStringValue("version", v.Version).
		WithStringValue("date", v.Date).
		WithStringValue("commit", v.Commit).
		Msg("Aserto policy CLI.")

	g.App.Cancel()
	<-g.App.Context.Done()

	return nil
}
