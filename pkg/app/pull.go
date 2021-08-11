package app

import (
	"github.com/containerd/containerd/remotes/docker"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

func (c *PolicyApp) Pull(userRef string) error {
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

	c.UI.Normal().
		WithStringValue("ref", userRef).
		Msg("Pulling.")

	resolver := docker.NewResolver(docker.ResolverOptions{})
	allowedMediaTypes := []string{MediaTypeImageLayer}
	descriptor, _, err := oras.Pull(c.Context, resolver, ref, ociStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return err
	}

	ociStore.AddReference(ref, descriptor)
	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	c.UI.Normal().
		WithStringValue("digest", descriptor.Digest.String()).
		Msgf("Pulled ref [%s].", ref)

	return nil
}
