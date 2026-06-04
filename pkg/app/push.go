package app

import (
	"github.com/opcr-io/policy/internal/oci"
	"github.com/opcr-io/policy/internal/parser"
	"github.com/opcr-io/policy/pkg/errors"
)

func (c *PolicyApp) Push(userRef string) error {
	defer c.Cancel()

	ref, err := parser.CalculateRef(userRef, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.ErrPushFailed.WithError(err)
	}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return errors.ErrPushFailed.WithError(err)
	}

	refs, err := ociClient.ListReferences()
	if err != nil {
		return errors.ErrPushFailed.WithError(err)
	}

	refDescriptor, ok := refs[ref]
	if !ok {
		return errors.ErrNotFound.WithMessage("policy [%s] not in the local store", ref)
	}

	c.UI.Normal().
		WithStringValue("digest", refDescriptor.Digest.String()).
		Msgf("Resolved ref [%s].", ref)

	digest, err := ociClient.Push(ref)
	if err != nil {
		return errors.ErrPushFailed.WithError(err)
	}

	c.UI.Normal().
		WithStringValue("digest", digest.String()).
		Msgf("Pushed ref [%s].", ref)

	return nil
}
