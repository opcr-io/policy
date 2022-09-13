package oci

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"strings"

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
	ociStore  *content.OCI
}

func NewOCI(ctx context.Context, log *zerolog.Logger, hostsFunc docker.RegistryHosts, policyRoot string) (*Oci, error) {
	ociStore, err := content.NewOCI(policyRoot)
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

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})

	var layers []ocispec.Descriptor
	allowedMediaTypes := []string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/octet-stream",
		"application/vnd.oci.image.config.v1+json",
		"application/vnd.unknown.config.v1+json",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar",
	}
	opts := []oras.CopyOpt{
		oras.WithAllowedMediaTypes(allowedMediaTypes),
		oras.WithAdditionalCachedMediaTypes(allowedMediaTypes...),
		oras.WithLayerDescriptors(func(d []ocispec.Descriptor) { layers = d }),
	}
	_, err := oras.Copy(o.ctx, resolver, ref, o.ociStore, "", opts...)
	if err != nil {
		return "", errors.Wrap(err, "oras pull failed")
	}

	var layer ocispec.Descriptor
	for _, lyr := range layers {
		if strings.Contains(lyr.MediaType, "tar") {
			layer = lyr
			break
		}
	}

	o.ociStore.AddReference(ref, layer)
	err = o.ociStore.SaveIndex()
	if err != nil {
		return "", err
	}

	return layer.Digest, nil
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

	allowedMediaTypes := []string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/octet-stream",
		"application/vnd.oci.image.config.v1+json",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar",
	}

	opts := []oras.CopyOpt{oras.WithAllowedMediaTypes(allowedMediaTypes)}

	memoryStore := content.NewMemory()

	config, configDescriptor, err := content.GenerateConfig(nil)
	if err != nil {
		return "", err
	}
	configDescriptor.MediaType = MediaTypeConfig
	manifest, manifestdesc, err := content.GenerateManifest(&configDescriptor, refDescriptor.Annotations, refDescriptor)
	if err != nil {
		return "", err
	}

	err = memoryStore.StoreManifest(ref, manifestdesc, manifest)
	if err != nil {
		return "", err
	}

	memoryStore.Set(configDescriptor, config)

	pushDescriptor, err := oras.Copy(o.ctx,
		o.ociStore,
		ref,
		resolver,
		"",
		opts...)

	if err != nil {
		return "", errors.Wrap(err, "oras push tarball failed")
	}

	// v1 version of oras-go doesn't push the manifest automatically so this part handles manifest pushing
	pushDescriptor, err = oras.Copy(o.ctx,
		memoryStore,
		ref,
		resolver,
		"",
		opts...)

	if err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
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

func CopyPolicy(ctx context.Context, log *zerolog.Logger,
	sourceRef, sourceUser, sourcePassword,
	destinationRef, destinationUser, destinationPassword,
	ociStore string) error {

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
	err = json.Unmarshal(b, &result)
	if err != nil {
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
