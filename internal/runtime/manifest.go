package runtime

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/open-policy-agent/opa/v1/bundle"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Manifest struct {
	bundle.Manifest

	RequiredBuiltIns StubBuiltinDefs
}

type MetadataEx struct {
	Metadata struct {
		RequiredBuiltins *StubBuiltinDefs `json:"required_builtins"`
	} `json:"metadata"`
}

func NewManifest() *Manifest {
	return &Manifest{
		Manifest:         bundle.Manifest{},
		RequiredBuiltIns: StubBuiltinDefs{},
	}
}

func LoadManifest(path string) (*Manifest, error) {
	m := NewManifest()

	switch fi, err := os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return m, status.Errorf(codes.NotFound, "%q file not found", path)
	case err != nil:
		return m, err
	case fi.IsDir():
		return m, status.Errorf(codes.NotFound, "%q is a directory", path)
	default:
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}

	manifest := bundle.Manifest{}
	if err := json.Unmarshal(buf, &manifest); err != nil {
		return m, err
	}

	metadata := MetadataEx{}
	if err := json.Unmarshal(buf, &metadata); err != nil {
		return m, err
	}

	required := StubBuiltinDefs{}
	if metadata.Metadata.RequiredBuiltins != nil {
		required = *metadata.Metadata.RequiredBuiltins
	}

	return &Manifest{
		Manifest:         manifest,
		RequiredBuiltIns: required,
	}, nil
}
