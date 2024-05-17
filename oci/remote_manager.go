//nolint:gocritic // Big parameter linter error passing ocispec.Descriptor, needed to implement oras.ReadOnlyTarget interface.
package oci

import (
	"bufio"
	"context"
	"io"
	"strings"

	"oras.land/oras-go/v2/content"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type remoteManager struct {
	resolver remotes.Resolver
	fetcher  content.Fetcher
	srcRef   string
}

func (r *remoteManager) Resolve(ctx context.Context, ref string) (ocispec.Descriptor, error) {
	_, desc, err := r.resolver.Resolve(ctx, ref)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

func (r *remoteManager) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	fetcher, err := r.resolver.Fetcher(ctx, r.srcRef)
	if err != nil {
		return nil, err
	}
	return fetcher.Fetch(ctx, target)
}

func (r *remoteManager) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	_, err := r.Fetch(ctx, target)
	if err == nil {
		return true, nil
	}

	return false, err
}

func (r *remoteManager) Push(ctx context.Context, expected ocispec.Descriptor, ctn io.Reader) error {
	pusher, err := r.resolver.Pusher(ctx, r.srcRef)
	if err != nil {
		return err
	}

	writer, err := pusher.Push(ctx, expected)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	defer func() {
		writer.Close()
	}()
	reader := bufio.NewReader(ctn)

	size, err := reader.WriteTo(writer)
	if err != nil {
		return err
	}
	return writer.Commit(ctx, size, expected.Digest)
}

func (r *remoteManager) Tag(ctx context.Context, desc ocispec.Descriptor, reference string) error {
	originalRef := r.srcRef
	reader, err := r.fetcher.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	r.srcRef = reference
	desc.Annotations = make(map[string]string)
	desc.Annotations[ocispec.AnnotationRefName] = reference
	err = r.Push(ctx, desc, reader)
	if err != nil {
		return err
	}

	r.srcRef = originalRef

	return nil
}

func (r *remoteManager) Delete(ctx context.Context, desc ocispec.Descriptor) error {
	return nil
}
