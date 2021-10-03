package main

import "github.com/opcr-io/policy/pkg/version"

type VersionCmd struct {
}

func (c *VersionCmd) Run(g *Globals) error {
	v := version.GetInfo()

	g.App.UI.Normal().
		WithStringValue("version", v.Version).
		WithStringValue("date", v.Date).
		WithStringValue("commit", v.Commit).
		Msg("Policy CLI.")

	g.App.Cancel()
	<-g.App.Context.Done()

	return nil
}
