package app

import (
	"os"
	"sort"
	"strings"

	"github.com/distribution/reference"
	"github.com/dustin/go-humanize"
	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/pkg/table"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type imageStruct struct {
	familiarName string
	tagOrNone    string
	digest       string
	createdAt    string
	size         string
}

//nolint:funlen
func (c *PolicyApp) Images() error {
	defer c.Cancel()

	var images []imageStruct

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

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

		var manifest *v1.Manifest
		if desc.MediaType == v1.MediaTypeImageManifest {
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
			createdAt = desc.Annotations[v1.AnnotationCreated]
		} else {
			createdAt = manifest.Annotations[v1.AnnotationCreated]
		}

		images = append(images, imageStruct{
			familiarName: familiarName,
			tagOrNone:    tagOrNone,
			digest:       desc.Digest.Encoded()[:12],
			createdAt:    createdAt,
			size:         strings.ReplaceAll(humanize.Bytes(uint64(desc.Size)), " ", ""), //nolint: gosec
		})
	}

	// sort data by CreatedAt DESC.
	sort.SliceStable(images, func(i, j int) bool {
		return images[i].createdAt < images[j].createdAt || (images[i].createdAt == images[j].createdAt && images[i].familiarName < images[j].familiarName)
	})

	data := [][]any{}
	for i := len(images) - 1; i >= 0; i-- {
		data = append(data, []any{
			images[i].familiarName,
			images[i].tagOrNone,
			images[i].digest,
			images[i].createdAt,
			images[i].size,
		})
	}

	t := table.New(os.Stdout)
	t.Header("Repository", "Tag", "Image ID", "Created", "Size")
	t.Bulk(data)
	t.Render()

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
