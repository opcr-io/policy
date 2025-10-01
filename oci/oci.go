package oci

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/v2/core/remotes/docker"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/content/oci"
)

const (
	MediaTypeImageLayer = "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaTypeConfig     = "application/vnd.oci.image.config.v1+json"
)

type Oci struct {
	logger         *zerolog.Logger
	ctx            context.Context
	hostsFunc      docker.RegistryHosts
	ociStore       *oci.Store
	policyRootPath string
}

func NewOCI(ctx context.Context, log *zerolog.Logger, hostsFunc docker.RegistryHosts, policyRoot string) (*Oci, error) {
	ociStore, err := oci.New(policyRoot)
	if err != nil {
		return nil, err
	}

	ociStore.AutoSaveIndex = true

	return &Oci{
		logger:         log,
		ctx:            ctx,
		hostsFunc:      hostsFunc,
		ociStore:       ociStore,
		policyRootPath: policyRoot,
	}, nil
}

func (o *Oci) WithDockerRegistryHosts(transport *http.Transport, user, pat string) *Oci {
	client := &http.Client{Transport: transport}

	authOpts := []docker.AuthorizerOpt{
		docker.WithAuthClient(client),
	}
	if user != "" || pat != "" {
		authOpts = append(
			authOpts,
			docker.WithAuthCreds(func(s string) (string, string, error) {
				return user, pat, nil
			}))
	}

	o.hostsFunc = func(host string) ([]docker.RegistryHost, error) {
		config := docker.RegistryHost{
			Host:         host,
			Scheme:       "https",
			Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
			Client:       client,
			Path:         "/v2",
			Authorizer:   docker.NewDockerAuthorizer(authOpts...),
		}

		return []docker.RegistryHost{config}, nil
	}

	return o
}

func (o *Oci) Pull(ref string) (digest.Digest, error) {
	dockerResolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})
	remoteManager := &remoteManager{resolver: dockerResolver, srcRef: ref}

	var manifestDescriptor ocispec.Descriptor

	opts := oras.DefaultCopyOptions

	// Get tarball descriptor digest
	opts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		if !IsAllowedMediaType(desc.MediaType) {
			return errors.Errorf("%s media type not allowed", desc.MediaType)
		}

		if strings.Contains(desc.MediaType, "manifest") {
			manifestDescriptor = desc
		}

		return nil
	}

	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		if !IsAllowedMediaType(desc.MediaType) {
			return errors.Errorf("%s media type not allowed", desc.MediaType)
		}

		if strings.Contains(desc.MediaType, "manifest") {
			manifestDescriptor = desc
		}

		return nil
	}

	if _, err := oras.Copy(o.ctx, remoteManager, ref, o.ociStore, "", opts); err != nil {
		return "", errors.Wrap(err, "oras pull failed")
	}

	if len(manifestDescriptor.Digest) > 0 {
		if err := o.ociStore.Tag(o.ctx, manifestDescriptor, ref); err != nil {
			return "", err
		}
	}

	if err := o.ociStore.SaveIndex(); err != nil {
		return "", err
	}

	return manifestDescriptor.Digest, nil
}

func (o *Oci) ListReferences() (map[string]ocispec.Descriptor, error) {
	var tgs []string

	refs := make(map[string]ocispec.Descriptor, 0)

	if err := o.ociStore.Tags(o.ctx, "", func(tags []string) error {
		tgs = append(tgs, tags...)
		return nil
	}); err != nil {
		return nil, err
	}

	for _, tag := range tgs {
		descr, err := o.ociStore.Resolve(o.ctx, tag)
		if err != nil {
			return nil, err
		}

		refs[tag] = descr
	}

	return refs, nil
}

func (o *Oci) Push(ref string) (digest.Digest, error) {
	dockerResolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})
	remoteManager := &remoteManager{resolver: dockerResolver, srcRef: ref, fetcher: o.ociStore}

	descriptor, err := o.ociStore.Resolve(o.ctx, ref)
	if err != nil {
		return "", err
	}

	if descriptor.MediaType == MediaTypeImageLayer {
		return o.pushBasedOnTarBall(remoteManager, &descriptor, ref)
	}

	tarBallDescriptor, configDescriptor, err := o.GetTarballAndConfigLayerDescriptor(o.ctx, &descriptor)
	if err != nil {
		return "", err
	}

	tarBallDescriptor.MediaType = MediaTypeImageLayer
	configDescriptor.MediaType = MediaTypeConfig

	// remove manifest from index
	err = o.ociStore.Untag(o.ctx, ref)
	if err != nil {
		return "", err
	}

	// tag tarball
	err = o.ociStore.Tag(o.ctx, *tarBallDescriptor, ref)
	if err != nil {
		return "", err
	}

	// copy tarball to remote first
	_, err = oras.Copy(o.ctx, o.ociStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push tarball failed")
	}

	// remove tarball from index
	err = o.ociStore.Untag(o.ctx, ref)
	if err != nil {
		return "", err
	}

	// tag config
	err = o.ociStore.Tag(o.ctx, *configDescriptor, ref)
	if err != nil {
		return "", err
	}

	// copy config to remote
	_, err = oras.Copy(o.ctx, o.ociStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push config failed")
	}

	// remove config from index
	err = o.ociStore.Untag(o.ctx, ref)
	if err != nil {
		return "", err
	}

	// tag manifest
	err = o.ociStore.Tag(o.ctx, descriptor, ref)
	if err != nil {
		return "", err
	}

	// copy manifest to remote
	_, err = oras.Copy(o.ctx, o.ociStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
	}

	return descriptor.Digest, nil
}

func (o *Oci) Tag(existingRef, newRef string) error {
	refs, err := o.ListReferences()
	if err != nil {
		return errors.Wrap(err, "failed to list references")
	}

	descriptor, ok := refs[existingRef]
	if !ok {
		for _, v := range refs {
			if strings.HasPrefix(v.Digest.String(), "sha256:"+existingRef) {
				descriptor = v
				break
			}
		}

		if descriptor.Size == 0 {
			return errors.Errorf("policy [%s] not found in the local store", existingRef)
		}
	}

	_, err = cloneDescriptor(&descriptor)
	if err != nil {
		return err
	}

	err = o.ociStore.Tag(o.ctx, descriptor, newRef)
	if err != nil {
		return err
	}

	return o.ociStore.SaveIndex()
}

func (o *Oci) Untag(descr *ocispec.Descriptor, ref string) error {
	return o.ociStore.Untag(o.ctx, ref)
}

func (o *Oci) GetStore() *oci.Store {
	return o.ociStore
}

func (o *Oci) GetTarballAndConfigLayerDescriptor(
	ctx context.Context,
	descriptor *ocispec.Descriptor,
) (
	*ocispec.Descriptor,
	*ocispec.Descriptor,
	error,
) {
	if descriptor == nil {
		return nil, nil, errors.New("nil descriptor provided")
	}

	if descriptor.MediaType != ocispec.MediaTypeImageManifest {
		return nil, nil, errors.New("provided descriptor is not a manifest descriptor")
	}

	manifest, err := o.GetManifest(descriptor)
	if err != nil {
		return nil, nil, err
	}

	configDigest := manifest.Config.Digest

	configDescriptor, err := o.ociStore.Resolve(ctx, configDigest.String())
	if err != nil {
		return nil, nil, err
	}

	for _, layer := range manifest.Layers {
		if layer.MediaType == ocispec.MediaTypeImageLayerGzip || layer.MediaType == ocispec.MediaTypeImageLayer {
			tarballDescriptor, err := o.ociStore.Resolve(ctx, layer.Digest.String())
			if err != nil {
				return nil, nil, err
			}

			return &tarballDescriptor, &configDescriptor, nil
		}
	}

	return nil, nil, errors.New("could not find tarball and config descriptors")
}

func (o *Oci) GetManifest(descriptor *ocispec.Descriptor) (*ocispec.Manifest, error) {
	reader, err := o.GetStore().Fetch(o.ctx, *descriptor)
	if err != nil {
		return nil, err
	}

	manifestBytes := new(bytes.Buffer)
	if _, err = manifestBytes.ReadFrom(reader); err != nil {
		return nil, err
	}

	var manifest ocispec.Manifest

	err = json.Unmarshal(manifestBytes.Bytes(), &manifest)
	if err != nil {
		return nil, err
	}

	err = reader.Close()
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (o *Oci) pushBasedOnTarBall(remoteManager *remoteManager, desc *ocispec.Descriptor, ref string) (digest.Digest, error) {
	memoryStore := memory.New()
	configBytes := []byte("{}")
	configDesc := content.NewDescriptorFromBytes(MediaTypeConfig, configBytes)

	err := memoryStore.Push(o.ctx, configDesc, bytes.NewReader(configBytes))
	if err != nil {
		return "", err
	}

	//nolint:staticcheck
	manifestDesc, err := oras.Pack(o.ctx, memoryStore, MediaTypeConfig, []ocispec.Descriptor{*desc}, oras.PackOptions{
		PackImageManifest:   true,
		ManifestAnnotations: desc.Annotations,
		ConfigDescriptor:    &configDesc,
	})
	if err != nil {
		return "", err
	}

	if _, err := oras.Copy(o.ctx, o.ociStore, ref, remoteManager, "", oras.DefaultCopyOptions); err != nil {
		return "", errors.Wrap(err, "oras push failed")
	}

	if err := memoryStore.Tag(o.ctx, manifestDesc, ref); err != nil {
		return "", err
	}

	remoteManager.fetcher = memoryStore
	if _, err := oras.Copy(o.ctx, memoryStore, ref, remoteManager, "", oras.DefaultCopyOptions); err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
	}

	if err := memoryStore.Tag(o.ctx, configDesc, ref); err != nil {
		return "", err
	}

	if _, err := oras.Copy(o.ctx, memoryStore, ref, remoteManager, "", oras.DefaultCopyOptions); err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
	}

	return desc.Digest, nil
}

func CopyPolicy(
	ctx context.Context,
	log *zerolog.Logger,
	sourceRef, sourceUser, sourcePassword, destinationRef, destinationUser, destinationPassword, ociStore string,
) error {
	transport, err := getTransport(log)
	if err != nil {
		return errors.Wrap(err, "failed to create transport")
	}

	ociClient, err := NewOCI(ctx,
		log,
		func(server string) ([]docker.RegistryHost, error) {
			client := &http.Client{Transport: transport}

			return []docker.RegistryHost{
				{
					Host:         server,
					Scheme:       "https",
					Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
					Client:       client,
					Path:         "/v2",
					Authorizer: docker.NewDockerAuthorizer(
						docker.WithAuthClient(client),
						docker.WithAuthCreds(func(s string) (string, string, error) {
							return sourceUser, sourcePassword, nil
						})),
				},
			}, nil
		},
		ociStore)
	if err != nil {
		return errors.Wrap(err, "failed to create oci client")
	}

	_, err = ociClient.Pull(sourceRef)
	if err != nil {
		return errors.Wrap(err, "failed to pull image")
	}

	err = ociClient.Tag(sourceRef, destinationRef)
	if err != nil {
		return errors.Wrap(err, "failed to tag image")
	}

	ociClient, err = NewOCI(ctx,
		log,
		func(server string) ([]docker.RegistryHost, error) {
			client := &http.Client{Transport: transport}

			return []docker.RegistryHost{
				{
					Host:         server,
					Scheme:       "https",
					Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
					Client:       client,
					Path:         "/v2",
					Authorizer: docker.NewDockerAuthorizer(
						docker.WithAuthClient(client),
						docker.WithAuthCreds(func(s string) (string, string, error) {
							return destinationUser, destinationPassword, nil
						})),
				},
			}, nil
		},
		ociStore)
	if err != nil {
		return errors.Wrap(err, "failed to create oci client")
	}

	_, err = ociClient.Push(destinationRef)
	if err != nil {
		return errors.Wrap(err, "failed to push image")
	}

	return nil
}

func cloneDescriptor(desc *ocispec.Descriptor) (ocispec.Descriptor, error) {
	b, err := json.Marshal(desc)
	if err != nil {
		return ocispec.Descriptor{}, errors.Wrap(err, "failed to clone descriptor")
	}

	result := ocispec.Descriptor{}

	if err := json.Unmarshal(b, &result); err != nil {
		return ocispec.Descriptor{}, errors.Wrap(err, "failed to create descriptor clone")
	}

	return result, nil
}

func getTransport(log *zerolog.Logger) (*http.Transport, error) {
	// Get the SystemCertPool, continue with an empty pool on error
	var rootCAs *x509.CertPool

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load system cert pool")
	}

	if rootCAs == nil {
		log.Warn().Err(err).Msg("failed to load system ca certs")

		rootCAs = x509.NewCertPool()
	}

	// Trust the augmented cert pool in our client
	conf := &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}

	return &http.Transport{TLSClientConfig: conf}, nil
}

func IsAllowedMediaType(mediatype string) bool {
	allowedMediaTypes := []string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/octet-stream",
		"application/vnd.oci.image.config.v1+json",
		"application/vnd.unknown.config.v1+json",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar",
	}

	for i := range allowedMediaTypes {
		if allowedMediaTypes[i] == mediatype {
			return true
		}
	}

	return false
}

func RemoveBlob(ref *ocispec.Descriptor, path string) error {
	digestPath := filepath.Join(strings.Split(ref.Digest.String(), ":")...)
	blob := filepath.Join(path, "blobs", digestPath)

	if err := os.Remove(blob); err != nil {
		return err
	}

	return nil
}
