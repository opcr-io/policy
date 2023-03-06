package app

import (
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Rm(existingRef string, force bool) error {
	defer c.Cancel()

	existingRefParsed, err := parser.CalculatePolicyRef(existingRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	confirmation := force
	if !force {
		c.UI.Exclamation().
			WithStringValue("reference", existingRefParsed).
			WithAskBoolMap("[Y/n]", &confirmation, map[string]bool{
				"":  true,
				"y": true,
				"n": false,
			}).Msgf("Are you sure?")
	}

	if !confirmation {
		c.UI.Exclamation().Msg("Operation canceled by user.")
		return nil
	}

	ociStore, err := content.NewOCI(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()

	_, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", existingRef)
	}

	ociStore.DeleteReference(existingRefParsed)

	// TODO: if there are no references left to the policy, perhaps delete the descriptor?
	// or implement a cleanup command.

	c.UI.Normal().
		WithStringValue("reference", existingRef).
		Msg("Removed reference.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}
