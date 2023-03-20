package app

import (
	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

	if contentInfo.MediaType == ocispec.MediaTypeImageManifest {
		manifest, err := ociClient.GetManifest(&contentInfo)
		if err != nil {
			return err
		}
		if len(manifest.Annotations) > 0 {
			msg := c.UI.Normal().WithTable("Annotation", "Value")

			for k, v := range manifest.Annotations {
				msg.WithTableRow(k, v)
			}
			msg.Msg("Annotations")
		}

	} else if len(contentInfo.Annotations) > 0 {
		msg := c.UI.Normal().WithTable("Annotation", "Value")

		for k, v := range contentInfo.Annotations {
			msg.WithTableRow(k, v)
		}
		msg.Msg("Annotations")
	}

	return nil
}
