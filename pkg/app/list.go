package app

import (
	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) List() error {
	defer func() {
		c.Cancel()
	}()

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
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

		name := ref.Name()

		tagOrDigest := ""
		tag, okTag := ref.(docker.Tagged)
		if okTag {
			tagOrDigest = tag.Tag()
		} else {
			tagOrDigest = info.Digest.String()
		}

		familiarName, err := c.familiarPolicyRef(name)
		if err != nil {
			return err
		}

		table.WithTableRow(familiarName, tagOrDigest, humanize.Bytes(uint64(v.Size)), humanize.Time(info.CreatedAt))
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
