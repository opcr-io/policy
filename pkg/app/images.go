package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Images() error {
	defer c.Cancel()

	var data [][]string

	ociStore, err := content.NewOCI(c.Configuration.PoliciesRoot())
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

	// sort data by CreatedAt DESC.
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

func (c *PolicyApp) ImagesRemote(server, org string, showEmpty bool) error {
	defer c.Cancel()

	creds := c.Configuration.Servers[server]

	xClient, err := extendedregistry.GetExtendedClient(
		c.Context,
		server,
		c.Logger,
		&extendedregistry.Config{
			Address:  "https://" + server,
			Username: creds.Username,
			Password: creds.Password,
		},
		c.TransportWithTrustedCAs())

	if err != nil {
		c.Logger.Debug().Err(err).Msgf("failed to get extended client for %s", server)
		c.UI.Exclamation().Msg("The registry doesn't support extended capabilities like listing policies.")
		return nil
	}

	p := c.UI.Progress("Fetching tags for images")
	p.Start()

	response, err := c.listImages(xClient, org)
	if err != nil {
		return err
	}

	imageData := [][]string{}
	for _, image := range response.Images {
		repo := fmt.Sprintf("%s/%s", server, image.Name)
		tags, _, err := xClient.ListTags(c.Context, org, image.Name, &api.PaginationRequest{Size: -1, Token: ""}, true)
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
			imageData = append(imageData, []string{familiarName, tag.Name, publicMark})
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

	// Get a list of tags for each image.
	return nil
}

// TODO: Expose pagination options.
func (c *PolicyApp) listImages(xClient extendedregistry.ExtendedClient, org string) (*registry.ListImagesResponse, error) {
	var response *registry.ListImagesResponse

	response, _, err := xClient.ListRepos(c.Context, org, &api.PaginationRequest{Size: -1, Token: ""})
	if err != nil {
		// TODO: use cerr.
		initialError := errors.Cause(err)
		if !strings.Contains(initialError.Error(), "authentication failed") {
			return nil, err
		}
	}
	if response == nil {
		response = &registry.ListImagesResponse{}
		response.Images = make([]*api.PolicyImage, 0)
	}

	if len(response.Images) != 0 {
		return response, nil
	}

	responsePublic, err := xClient.ListPublicRepos(c.Context, org, &api.PaginationRequest{Size: -1, Token: ""})
	if err != nil {
		return nil, err
	}

	for _, image := range responsePublic.Images {
		response.Images = append(response.Images, &api.PolicyImage{Name: fmt.Sprintf("%s/%s", org, image.Name), Public: image.Public})
	}

	return response, nil
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
