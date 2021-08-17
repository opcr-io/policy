package app

import (
	"github.com/containerd/containerd/remotes/docker"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

func (c *PolicyApp) Pull(userRef string) error {
	defer c.Cancel()

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

	c.UI.Normal().
		WithStringValue("ref", userRef).
		Msg("Pulling.")

	resolver := docker.NewResolver(
		docker.ResolverOptions{
			Credentials: func(s string) (string, string, error) {
				serverCreds, ok := c.Configuration.Servers[s]
				if !ok {
					return "", "", nil
				}

				return serverCreds.Username, serverCreds.Password, nil
			},
		},
	)
	allowedMediaTypes := []string{MediaTypeImageLayer}
	opts := []oras.PullOpt{
		oras.WithAllowedMediaTypes(allowedMediaTypes),
		oras.WithContentProvideIngester(ociStore),
	}
	_, descriptors, err := oras.Pull(c.Context, resolver, ref, ociStore,
		opts...,
	)
	if err != nil {
		return errors.Wrap(err, "oras pull failed")
	}

	if len(descriptors) != 1 {
		return errors.Errorf("unexpected layer count of [%d] from the registry; expected 1", len(descriptors))
	}

	ociStore.AddReference(ref, descriptors[0])
	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	c.UI.Normal().
		WithStringValue("digest", descriptors[0].Digest.String()).
		Msgf("Pulled ref [%s].", ref)

	return nil
}
