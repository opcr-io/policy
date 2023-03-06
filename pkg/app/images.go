package app

import (
	"sort"
	"strings"
	"time"

	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
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
