package app

import (
	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Inspect(userRef string) error {
	defer c.Cancel()

	ref, err := parser.CalculatePolicyRef(userRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	contentInfo, err := ociClient.GetStore().Resolve(c.Context, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to read content info for policy [%s]", ref)
	}

	c.UI.Normal().
		WithStringValue("media type", contentInfo.MediaType).
		WithStringValue("digest", contentInfo.Digest.String()).
		WithIntValue("size", contentInfo.Size).
		Do()

	if len(contentInfo.Annotations) > 0 {
		msg := c.UI.Normal().WithTable("Annotation", "Value")

		for k, v := range contentInfo.Annotations {
			msg.WithTableRow(k, v)
		}
		msg.Msg("Annotations")
	}

	return nil
}
