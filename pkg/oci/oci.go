package oci

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
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
	logger    *zerolog.Logger
	ctx       context.Context
	hostsFunc docker.RegistryHosts
	ociStore  *oci.Store
}

func NewOCI(ctx context.Context, log *zerolog.Logger, hostsFunc docker.RegistryHosts, policyRoot string) (*Oci, error) {
	ociStore, err := oci.New(policyRoot)
	if err != nil {
		return nil, err
	}
	ociStore.AutoSaveIndex = true

	return &Oci{
		logger:    log,
		ctx:       ctx,
		hostsFunc: hostsFunc,
		ociStore:  ociStore,
	}, nil
}

func (o *Oci) Pull(ref string) (digest.Digest, error) {
	dockerResolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.hostsFunc,
	})
	remoteManager := &remoteManager{resolver: dockerResolver, srcRef: ref}

	// var layers []ocispec.Descriptor
	// allowedMediaTypes := []string{
	// 	"application/vnd.oci.image.manifest.v1+json",
	// 	"application/octet-stream",
	// 	"application/vnd.oci.image.config.v1+json",
	// 	"application/vnd.unknown.config.v1+json",
	// 	"application/vnd.oci.image.layer.v1.tar+gzip",
	// 	"application/vnd.oci.image.layer.v1.tar",
	// }
	// opts := []oras.CopyOpt{
	// 	oras.WithAllowedMediaTypes(allowedMediaTypes),
	// 	oras.WithAdditionalCachedMediaTypes(allowedMediaTypes...),
	// 	oras.WithLayerDescriptors(func(d []ocispec.Descriptor) { layers = d }),
	// }
	descr, err := oras.Copy(o.ctx, remoteManager, ref, o.ociStore, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras pull failed")
	}

	// var layer ocispec.Descriptor
	// for _, lyr := range layers {
	// 	if strings.Contains(lyr.MediaType, "tar") {
	// 		layer = lyr
	// 		break
	// 	}
	// }

	// o.ociStore.AddReference(ref, layer)
	err = o.ociStore.SaveIndex()
	if err != nil {
		return "", err
	}

	return descr.Digest, nil
}

func (o *Oci) ListReferences() (map[string]ocispec.Descriptor, error) {
	var tgs []string
	refs := make(map[string]ocispec.Descriptor, 0)
	err := o.ociStore.Tags(o.ctx, "", func(tags []string) error {
		tgs = append(tgs, tags...)
		return nil
	})

	if err != nil {
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

	desc, err := o.ociStore.Resolve(o.ctx, ref)
	if err != nil {
		return "", err
	}

	memoryStore := memory.New()
	configBytes := []byte("{}")
	configDesc := content.NewDescriptorFromBytes(MediaTypeConfig, configBytes)

	err = memoryStore.Push(o.ctx, configDesc, bytes.NewReader(configBytes))
	if err != nil {
		return "", err
	}

	manifestDesc, err := oras.Pack(o.ctx, memoryStore, MediaTypeConfig, []ocispec.Descriptor{desc}, oras.PackOptions{
		PackImageManifest:   true,
		ManifestAnnotations: desc.Annotations,
		ConfigDescriptor:    &configDesc,
	})
	if err != nil {
		return "", err
	}

	descr, err := oras.Copy(o.ctx, o.ociStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push failed")
	}

	err = memoryStore.Tag(o.ctx, manifestDesc, ref)
	if err != nil {
		return "", err
	}

	remoteManager.fetcher = memoryStore
	_, err = oras.Copy(o.ctx, memoryStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
	}

	err = memoryStore.Tag(o.ctx, configDesc, ref)
	if err != nil {
		return "", err
	}

	_, err = oras.Copy(o.ctx, memoryStore, ref, remoteManager, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "oras push manifest failed")
	}

	// refs, err := o.ListReferences()
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to list references")
	// }

	// refDescriptor, ok := refs[ref]
	// if !ok {
	// 	return "", errors.Errorf("policy [%s] not found in the local store", ref)

	// }

	// resolver := docker.NewResolver(docker.ResolverOptions{
	// 	Hosts: o.hostsFunc,
	// })

	// delete(refDescriptor.Annotations, "org.opencontainers.image.ref.name")

	// allowedMediaTypes := []string{
	// 	"application/vnd.oci.image.manifest.v1+json",
	// 	"application/octet-stream",
	// 	"application/vnd.oci.image.config.v1+json",
	// 	"application/vnd.oci.image.layer.v1.tar+gzip",
	// 	"application/vnd.oci.image.layer.v1.tar",
	// }

	// opts := []oras.CopyOpt{oras.WithAllowedMediaTypes(allowedMediaTypes)}

	// memoryStore := content.NewMemory()

	// config, configDescriptor, err := content.GenerateConfig(nil)
	// if err != nil {
	// 	return "", err
	// }
	// configDescriptor.MediaType = MediaTypeConfig
	// manifest, manifestdesc, err := content.GenerateManifest(&configDescriptor, refDescriptor.Annotations, refDescriptor)
	// if err != nil {
	// 	return "", err
	// }

	// err = memoryStore.StoreManifest(ref, manifestdesc, manifest)
	// if err != nil {
	// 	return "", err
	// }

	// memoryStore.Set(configDescriptor, config)

	// pushDescriptor, err := oras.Copy(o.ctx,
	// 	o.ociStore,
	// 	ref,
	// 	resolver,
	// 	"",
	// 	opts...)

	// if err != nil {
	// 	return "", errors.Wrap(err, "oras push tarball failed")
	// }

	// // v1 version of oras-go doesn't push the manifest automatically so this part handles manifest pushing
	// pushDescriptor, err = oras.Copy(o.ctx,
	// 	memoryStore,
	// 	ref,
	// 	resolver,
	// 	"",
	// 	opts...)

	// if err != nil {
	// 	return "", errors.Wrap(err, "oras push manifest failed")
	// }

	// return pushDescriptor.Digest, nil
	return descr.Digest, nil
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

func (r *remoteManager) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	pusher, err := r.resolver.Pusher(ctx, r.srcRef)
	if err != nil {
		return err
	}

	writer, err := pusher.Push(ctx, expected)
	if err != nil {
		return err
	}

	defer func() {
		writer.Close()
	}()

	reader := bufio.NewReader(content)
	size, err := reader.WriteTo(writer)
	if err != nil {
		return err
	}

	return writer.Commit(ctx, size, expected.Digest)
}

func (r *remoteManager) Tag(ctx context.Context, desc ocispec.Descriptor, reference string) error {
	originalRef := r.srcRef
	content, err := r.fetcher.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	r.srcRef = reference
	err = r.Push(ctx, desc, content)
	if err != nil {
		return err
	}

	r.srcRef = originalRef
	return nil
}
