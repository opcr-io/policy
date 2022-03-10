package app

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Images() error {
	defer c.Cancel()

	var data [][]string

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return nil
	}

	table := c.UI.Normal().WithTable("Repository", "Tag", "Image ID", "Created", "Size")
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

		tagOrNone := "<none>"
		tag, okTag := ref.(docker.Tagged)
		if okTag {
			tagOrNone = tag.Tag()
		}

		familiarName, err := c.familiarPolicyRef(refName)
		if err != nil {
			return err
		}

		arrData := []string{
			familiarName,
			tagOrNone,
			info.Digest.Encoded()[:12],
			humanize.Time(info.CreatedAt),
			strings.ReplaceAll(humanize.Bytes(uint64(v.Size)), " ", ""),
			info.CreatedAt.Format(time.RFC3339Nano)}

		data = append(data, arrData)
	}

	// sort data by CreatedAt DESC
	sort.SliceStable(data, func(i, j int) bool {
		return data[i][5] < data[j][5] || (data[i][5] == data[j][5] && data[i][1] < data[j][1])
	})

	for i := len(data) - 1; i >= 0; i-- {
		v := data[i]
		table.WithTableRow(v[0], v[1], v[2], v[3], v[4])
	}

	table.Do()

	return nil
}

func (c *PolicyApp) ImagesRemote(server string, showEmpty bool) error {
	defer c.Cancel()

	creds := c.Configuration.Servers[server]

	xClient, err := extendedregistry.GetExtendedClient(server,
		c.Logger,
		&extendedregistry.Config{
			Address:  "https://" + server,
			Username: creds.Username,
			Password: creds.Password,
		},
		c.TransportWithTrustedCAs())

	if err != nil {
		c.UI.Exclamation().Msg("The registry doesn't support extended capabilities like listing policies.")
		return nil
	}

	// If the server doesn't support list APIs, print a message and return.
	// if !xClient.Supported() {
	// 	c.UI.Exclamation().Msg("The registry doesn't support extended capabilities like listing policies.")
	// 	return nil
	// }

	// Get a list of all images
	images, err := xClient.ListRepos()
	if err != nil {
		return err
	}

	p := c.UI.Progress("Fetching tags for images")
	p.Start()

	imageData := [][]string{}
	for _, image := range images {
		repo := server + "/" + image.Name

		tags, err := c.imageTags(repo, creds.Username, creds.Password)
		if err != nil {
			return err
		}

		familiarName, err := c.familiarPolicyRef(repo)
		if err != nil {
			return err
		}

		if len(tags) == 0 && showEmpty {
			imageData = append(imageData, []string{familiarName, "<no tags>", "-"})
			continue
		}

		publicMark := "No"
		if image.Public {
			publicMark = "Yes"
		}

		for _, tag := range tags {
			imageData = append(imageData, []string{familiarName, tag, publicMark})
		}
	}

	p.Stop()

	sort.SliceStable(imageData, func(i, j int) bool {
		return imageData[i][0] < imageData[j][0] || (imageData[i][0] == imageData[j][0] && imageData[i][1] < imageData[j][1])
	})

	table := c.UI.Normal().WithTable("Repository", "Tag", "Public")
	for _, image := range imageData {
		table.WithTableRow(image[0], image[1], image[2])
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
