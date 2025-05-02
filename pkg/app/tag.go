package app

import (
	"strings"

	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/parser"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Tag(existingRef, newRef string) error {
	defer c.Cancel()

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	existingRefParsed := existingRef
	if strings.Contains(existingRef, ":") || strings.Contains(existingRef, "/") {
		existingRefParsed, err = parser.CalculatePolicyRef(existingRef, c.Configuration.DefaultDomain)
		if err != nil {
			return err
		}
	}

	parsed, err := parser.CalculatePolicyRef(newRef, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	c.UI.Normal().
		WithStringValue("reference", newRef).
		Msg("Tagging image.")

	if err := ociClient.Tag(existingRefParsed, parsed); err != nil {
		return err
	}

	return nil
}
