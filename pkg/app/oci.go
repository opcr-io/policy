package app

import (
	"encoding/json"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

const (
	MediaTypeImageLayer    = "application/vnd.oci.image.layer.v1.tar+gzip"
	DefaultCanonicalDomain = "docker.io"
)

func cloneDescriptor(desc *v1.Descriptor) (v1.Descriptor, error) {
	b, err := json.Marshal(desc)
	if err != nil {
		return v1.Descriptor{}, errors.Wrap(err, "failed to clone descriptor")
	}

	result := v1.Descriptor{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return v1.Descriptor{}, errors.Wrap(err, "failed to create descriptor clone")
	}

	return result, nil
}
