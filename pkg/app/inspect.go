package app

import (
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Inspect(userRef string) error {
	defer c.Cancel()

	ref, err := c.calculatePolicyRef(userRef)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	refs := ociStore.ListReferences()

	refDescriptor, ok := refs[ref]
	if !ok {
		return errors.Errorf("policy [%s] not found in the local store", ref)
	}

	contentInfo, err := ociStore.Info(c.Context, refDescriptor.Digest)
	if err != nil {
		return errors.Wrapf(err, "failed to read content info for policy [%s]", ref)
	}

	c.UI.Normal().
		WithStringValue("digest", contentInfo.Digest.String()).
		WithIntValue("size", contentInfo.Size).
		WithStringValue("created_at", contentInfo.CreatedAt.String()).
		WithStringValue("updated_at", contentInfo.UpdatedAt.String()).
		Do()

	if len(refDescriptor.Annotations) > 0 {
		msg := c.UI.Normal().WithTable("Annotation", "Value")

		for k, v := range refDescriptor.Annotations {
			msg.WithTableRow(k, v)
		}

		msg.Msg("Annotations")
	}

	if len(contentInfo.Labels) > 0 {
		msg := c.UI.Normal().WithTable("Label", "Value")

		for k, v := range contentInfo.Labels {
			msg.WithTableRow(k, v)
		}

		msg.Msg("Labels")
	}

	return nil
}
