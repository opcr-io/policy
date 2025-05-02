package app

import (
	"sort"
	"strings"

	"github.com/distribution/reference"
	"github.com/dustin/go-humanize"
	"github.com/opcr-io/policy/oci"
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

	data := []imageStruct{}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	table := c.UI.Normal().WithTable("Repository", "Tag", "Image ID", "Created", "Size")

	var tgs []string

	if err := ociClient.GetStore().Tags(c.Context, "", func(tags []string) error {
		tgs = append(tgs, tags...)
		return nil
	}); err != nil {
		return err
	}

	for _, tag := range tgs {
		desc, err := ociClient.GetStore().Resolve(c.Context, tag)
		if err != nil {
			return err
		}

		var manifest *ocispec.Manifest
		if desc.MediaType == ocispec.MediaTypeImageManifest {
			manifest, err = ociClient.GetManifest(&desc)
			if err != nil {
				return err
			}
		}

		ref, err := reference.ParseDockerRef(tag)
		if err != nil {
			return err
		}

		refName := ref.Name()

		tagOrNone := "<none>"

		tag, okTag := ref.(reference.Tagged)
		if okTag {
			tagOrNone = tag.Tag()
		}

		familiarName, err := c.familiarPolicyRef(refName)
		if err != nil {
			return err
		}

		var createdAt string
		if manifest == nil {
			createdAt = desc.Annotations[ocispec.AnnotationCreated]
		} else {
			createdAt = manifest.Annotations[ocispec.AnnotationCreated]
		}

		data = append(data, imageStruct{
			familiarName: familiarName,
			tagOrNone:    tagOrNone,
			digest:       desc.Digest.Encoded()[:12],
			createdAt:    createdAt,
			size:         strings.ReplaceAll(humanize.Bytes(uint64(desc.Size)), " ", ""), //nolint: gosec
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
	parsed, err := reference.ParseDockerRef(userRef)
	if err != nil {
		return "", err
	}

	domain := reference.Domain(parsed)
	if domain == c.Configuration.DefaultDomain {
		path := reference.Path(parsed)
		return path, nil
	}

	return userRef, nil
}
