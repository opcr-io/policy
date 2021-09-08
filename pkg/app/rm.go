package app

import (
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Rm(existingRef string) error {
	defer c.Cancel()

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()
	existingRefParsed, err := c.calculatePolicyRef(existingRef)
	if err != nil {
		return err
	}

	_, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", existingRef)
	}

	ociStore.DeleteReference(existingRefParsed)

	// TODO: if there are no references left to the policy, perhaps delete the descriptor?
	// or implement a cleanup command

	c.UI.Normal().
		WithStringValue("reference", existingRef).
		Msg("Removed reference.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}
