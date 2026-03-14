package main

import (
	"os"

	"github.com/pkg/errors"
)

type untarDir string

func (d untarDir) AfterApply() error {
	if fi, err := os.Stat(string(d)); err == nil && fi.IsDir() {
		return nil
	}

	return errors.Errorf("--untardir directory %q does not exist", string(d))
}

type PullCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to pull from the remote registry."`
	UntarDir untarDir `flag:"" name:"untardir"  help:"Extract the policy bundle to an existing directory."`
}

func (c *PullCmd) Run(g *Globals) error {
	defer g.App.Context.Done()

	var errs error

	for _, policyRef := range c.Policies {
		if err := g.App.Pull(policyRef, string(c.UntarDir)); err != nil {
			g.App.UI.Problem().WithErr(err).Msgf("Failed to pull policy: %s", policyRef)
			errs = err
		}
	}

	if errs != nil {
		return errors.Wrap(errs, "failed to pull one or more policies")
	}

	return nil
}
