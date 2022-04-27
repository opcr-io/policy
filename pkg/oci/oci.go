package oci

import (
	"context"
	"encoding/json"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

const (
	// MediaTypeImageLayer = "application/vnd.opa.policy.v1.tar+gzip"
	MediaTypeImageLayer = "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaTypeConfig     = "application/vnd.oci.image.config.v1+json"
)

type Oci struct {
	logger    *zerolog.Logger
	ctx       context.Context
	hostsFunc docker.RegistryHosts
	ociStore  *content.OCIStore
}

func NewOCI(ctx context.Context, log *zerolog.Logger, hostsFunc docker.RegistryHosts, policyRoot string) (*Oci, error) {
	ociStore, err := content.NewOCIStore(policyRoot)
	if err != nil {
		return nil, err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return nil, err
	}

	return &Oci{
		logger:    log,
		ctx:       ctx,
		hostsFunc: hostsFunc,
		ociStore:  ociStore,
	}, nil
}

func (o *Oci) Pull(ref string) (digest.Digest, error) {
	opts := []oras.PullOpt{
		oras.WithContentProvideIngester(o.ociStore),
	}

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})

	_, descriptors, err := oras.Pull(o.ctx, resolver, ref, o.ociStore,
		opts...,
	)

	if err != nil {
		return "", errors.Wrap(err, "oras pull failed")
	}

	if len(descriptors) != 1 {
		return "", errors.Errorf("unexpected layer count of [%d] from the registry; expected 1", len(descriptors))
	}

	o.ociStore.AddReference(ref, descriptors[0])
	err = o.ociStore.SaveIndex()
	if err != nil {
		return "", err
	}
	return descriptors[0].Digest, nil
}

func (o *Oci) ListReferences() (map[string]ocispec.Descriptor, error) {
	refs := o.ociStore.ListReferences()
	return refs, nil
}

func (o *Oci) Push(ref string) (digest.Digest, error) {
	refs, err := o.ListReferences()
	if err != nil {
		return "", errors.Wrap(err, "failed to list references")
	}

	refDescriptor, ok := refs[ref]
	if !ok {
		return "", errors.Errorf("policy [%s] not found in the local store", ref)
	}

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})

	delete(refDescriptor.Annotations, "org.opencontainers.image.ref.name")

	pushDescriptor, err := oras.Push(o.ctx,
		resolver,
		ref,
		o.ociStore,
		[]ocispec.Descriptor{refDescriptor},
		oras.WithConfigMediaType("application/vnd.oci.image.config.v1+json"))

	if err != nil {
		return "", errors.Wrap(err, "oras push failed")
	}

	return pushDescriptor.Digest, nil
}

func (o *Oci) Tag(existingRef, newRef string) error {
	refs, err := o.ListReferences()
	if err != nil {
		return errors.Wrap(err, "failed to list references")
	}

	descriptor, ok := refs[existingRef]
	if !ok {
		return errors.Errorf("policy [%s] not found in the local store", existingRef)
	}

	newDescriptor, err := cloneDescriptor(&descriptor)
	if err != nil {
		return err
	}

	o.ociStore.AddReference(newRef, newDescriptor)

	return o.ociStore.SaveIndex()
}

func cloneDescriptor(desc *ocispec.Descriptor) (ocispec.Descriptor, error) {
	b, err := json.Marshal(desc)
	if err != nil {
		return ocispec.Descriptor{}, errors.Wrap(err, "failed to clone descriptor")
	}

	result := ocispec.Descriptor{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return ocispec.Descriptor{}, errors.Wrap(err, "failed to create descriptor clone")
	}

	return result, nil
}
