package app

import (
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Tag(existingRef, newRef string) error {
	defer c.Cancel()

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()
	existingRefParsed, err := parser.CalculatePolicyRef(existingRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	descriptor, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", existingRef)
	}

	parsed, err := parser.CalculatePolicyRef(newRef, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	newDescriptor, err := cloneDescriptor(&descriptor)
	if err != nil {
		return err
	}

	ociStore.AddReference(parsed, newDescriptor)

	c.UI.Normal().
		WithStringValue("reference", newRef).
		Msg("Tagging image.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}
