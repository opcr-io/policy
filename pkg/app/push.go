package app

import (
	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/parser"
	perr "github.com/opcr-io/policy/pkg/errors"
)

func (c *PolicyApp) Push(userRef string) error {
	defer c.Cancel()

	ref, err := parser.CalculatePolicyRef(userRef, c.Configuration.DefaultDomain)
	if err != nil {
		return perr.ErrPolicyPushFailed.WithError(err)
	}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return perr.ErrPolicyPushFailed.WithError(err)
	}

	refs, err := ociClient.ListReferences()
	if err != nil {
		return perr.ErrPolicyPushFailed.WithError(err)
	}

	refDescriptor, ok := refs[ref]
	if !ok {
		return perr.ErrPolicyNotFound.WithMessage("policy [%s] not in the local store", ref)
	}

	c.UI.Normal().
		WithStringValue("digest", refDescriptor.Digest.String()).
		Msgf("Resolved ref [%s].", ref)

	digest, err := ociClient.Push(ref)
	if err != nil {
		return perr.ErrPolicyPushFailed.WithError(err)
	}

	c.UI.Normal().
		WithStringValue("digest", digest.String()).
		Msgf("Pushed ref [%s].", ref)

	return nil
}
