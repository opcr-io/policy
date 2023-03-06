package app

import (
	"sort"
	"strings"

	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Images() error {
	defer c.Cancel()

	var data [][]string

	ociStore, err := oci.New(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	table := c.UI.Normal().WithTable("Repository", "Tag", "Image ID", "Size")
	var tgs []string
	ociStore.Tags(c.Context, "", func(tags []string) error {
		tgs = append(tgs, tags...)
		return nil
	})

	for _, tag := range tgs {
		descr, err := ociStore.Resolve(c.Context, tag)
		if err != nil {
			return err
		}

		ref, err := docker.ParseDockerRef(tag)
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

		fmt.Println(descr)

		arrData := []string{
			familiarName,
			tagOrNone,
			descr.Digest.Encoded()[:12],
			strings.ReplaceAll(humanize.Bytes(uint64(descr.Size)), " ", "")}

		data = append(data, arrData)
	}

	// sort data by CreatedAt DESC.
	sort.SliceStable(data, func(i, j int) bool {
		return data[i][3] < data[j][3] || (data[i][3] == data[j][3] && data[i][1] < data[j][1])
	})

	for i := len(data) - 1; i >= 0; i-- {
		v := data[i]
		table.WithTableRow(v[0], v[1], v[2], v[3])
	}

	table.Do()

	return nil
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
