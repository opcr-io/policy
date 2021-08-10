package app

import (
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

func (c *PolicyApp) Push(userRef string) error {
	defer func() {
		c.Cancel()
	}()

	ref, err := c.calculatePolicyRef(userRef)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
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
		c.UI.Exclamation().
			WithStringValue("policy", ref).
			Msgf("Policy reference not found in the local store.")

	}

	c.UI.Normal().
		WithStringValue("digest", refDescriptor.Digest.String()).
		Msgf("Resolved ref [%s].", ref)

	resolver := docker.NewResolver(docker.ResolverOptions{})

	refDescriptor.Annotations[ocispec.AnnotationTitle] = ref

	pushDescriptor, err := oras.Push(c.Context, resolver, ref, ociStore, []ocispec.Descriptor{refDescriptor})
	if err != nil {
		return err
	}

	c.UI.Normal().
		WithStringValue("digest", pushDescriptor.Digest.String()).
		Msgf("Pushed ref [%s].", ref)

	return nil
}
