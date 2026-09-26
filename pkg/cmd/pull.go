package cmd

import "github.com/pkg/errors"

type PullCmd struct {
	Policies []string `arg:"" name:"policy" help:"Policies to pull from the remote registry."`
	UntarDir string   `name:"untardir" help:"Extract the policy bundle to an existing directory." type:"existingdir" optional:""`
}

func (c *PullCmd) Run(g *Globals) error {
	var errs error

	for _, policyRef := range c.Policies {
		err := g.App.Pull(policyRef, c.UntarDir)
		if err != nil {
			if c.UntarDir != "" {
				g.App.UI.Problem().WithErr(err).Msgf("Failed to pull/extract policy: %s", policyRef)
			} else {
				g.App.UI.Problem().WithErr(err).Msgf("Failed to pull policy: %s", policyRef)
			}

			errs = err
		}
	}

	<-g.App.Context.Done()

	if errs != nil {
		return errors.Wrap(errs, "failed to pull one or more policies")
	}

	return nil
}
