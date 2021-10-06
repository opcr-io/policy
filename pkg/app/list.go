package app

import (
	"net/http"

	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	extendedclient "github.com/opcr-io/policy/pkg/extended_client"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Images() error {
	defer c.Cancel()

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return nil
	}

	table := c.UI.Normal().WithTable("Repository", "Tag", "Size", "Created At")
	refs := ociStore.ListReferences()
	for k, v := range refs {
		info, err := ociStore.Info(c.Context, v.Digest)
		if err != nil {
			return err
		}

		ref, err := docker.ParseDockerRef(k)
		if err != nil {
			return err
		}

		refName := ref.Name()

		tagOrDigest := ""
		tag, okTag := ref.(docker.Tagged)
		if okTag {
			tagOrDigest = tag.Tag()
		} else {
			tagOrDigest = info.Digest.String()
		}

		familiarName, err := c.familiarPolicyRef(refName)
		if err != nil {
			return err
		}

		table.WithTableRow(familiarName, tagOrDigest, humanize.Bytes(uint64(v.Size)), humanize.Time(info.CreatedAt))
	}
	table.Do()

	return nil
}

func (c *PolicyApp) ImagesRemote(server string) error {
	defer c.Cancel()

	creds := c.Configuration.Servers[server]

	xClient := extendedclient.NewExtendedClient(c.Logger, &extendedclient.Config{
		Address:  "https://" + server,
		Username: creds.Username,
		Password: creds.Password,
	},
		c.TransportWithTrustedCAs())

	// If the server doesn't support list APIs, print a message and return.
	if !xClient.Supported() {
		c.UI.Exclamation().Msg("The registry doesn't support extended capabilities like listing policies.")
		return nil
	}

	// Get a list of all images
	images, err := xClient.ListImages()
	if err != nil {
		return err
	}

	table := c.UI.Normal().WithTable("Repository", "Tag")
	for _, image := range images {
		repo := server + "/" + image

		tags, err := c.imageTags(repo, creds.Username, creds.Password)
		if err != nil {
			return err
		}

		familiarName, err := c.familiarPolicyRef(repo)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			table.WithTableRow(familiarName, tag)
		}
	}
	table.Do()

	// Get a list of tags for each image
	return nil
}

func (c *PolicyApp) imageTags(repoName, username, password string) ([]string, error) {
	repo, err := name.NewRepository(repoName)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid repo name [%s]", repoName)
	}

	tags, err := remote.List(repo,
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		}),
		remote.WithTransport(c.TransportWithTrustedCAs()))

	if err != nil {
		if tErr, ok := err.(*transport.Error); ok {
			switch tErr.StatusCode {
			case http.StatusUnauthorized:
				return nil, errors.Wrap(err, "authentication to docker registry failed")
			case http.StatusNotFound:
				return []string{}, nil
			}
		}

		return nil, errors.Wrap(err, "failed to list tags from registry")
	}

	return tags, nil
}

func (c *PolicyApp) familiarPolicyRef(userRef string) (string, error) {
	parsed, err := docker.ParseDockerRef(userRef)
	if err != nil {
		return "", err
	}

	domain := docker.Domain(parsed)
	if domain == c.Configuration.DefaultDomain {
		path := docker.Path(parsed)
		return path, nil
	}

	return userRef, nil
}
