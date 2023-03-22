package app

import (
	"sort"
	"strings"

	"github.com/containerd/containerd/reference/docker"
	"github.com/dustin/go-humanize"
	"github.com/opcr-io/policy/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type imageStruct struct {
	familiarName string
	tagOrNone    string
	digest       string
	createdAt    string
	size         string
}

func (c *PolicyApp) Images() error {
	defer c.Cancel()

	var data []imageStruct

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	table := c.UI.Normal().WithTable("Repository", "Tag", "Image ID", "Created", "Size")
	var tgs []string
	err = ociClient.GetStore().Tags(c.Context, "", func(tags []string) error {
		tgs = append(tgs, tags...)
		return nil
	})
	if err != nil {
		return err
	}

	for _, tag := range tgs {
		descr, err := ociClient.GetStore().Resolve(c.Context, tag)
		if err != nil {
			return err
		}
		var manifest *ocispec.Manifest
		if descr.MediaType == ocispec.MediaTypeImageManifest {
			manifest, err = ociClient.GetManifest(&descr)
			if err != nil {
				return err
			}
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
		var createdAt string
		if manifest == nil {
			createdAt = descr.Annotations[ocispec.AnnotationCreated]
		} else {
			createdAt = manifest.Annotations[ocispec.AnnotationCreated]
		}

		data = append(data, imageStruct{
			familiarName: familiarName,
			tagOrNone:    tagOrNone,
			digest:       descr.Digest.Encoded()[:12],
			createdAt:    createdAt,
			size:         strings.ReplaceAll(humanize.Bytes(uint64(descr.Size)), " ", ""),
		})
	}

	// sort data by CreatedAt DESC.
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].createdAt < data[j].createdAt || (data[i].createdAt == data[j].createdAt && data[i].familiarName < data[j].familiarName)
	})

	for i := len(data) - 1; i >= 0; i-- {
		table.WithTableRow(data[i].familiarName, data[i].tagOrNone, data[i].digest, data[i].createdAt, data[i].size)
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
