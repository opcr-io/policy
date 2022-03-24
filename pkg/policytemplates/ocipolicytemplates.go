package policytemplates

import (
	"compress/gzip"
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	tarfs "github.com/nlepage/go-tarfs"

	"github.com/containerd/containerd/remotes/docker"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

type Config struct {
	Server     string
	PolicyRoot string
}

type oci struct {
	logger    *zerolog.Logger
	extClient extendedregistry.ExtendedClient
	transport *http.Transport
	ctx       context.Context
	cfg       Config
}

// NewOCI returns a new policy template provider for OCI
func NewOCI(ctx context.Context, log *zerolog.Logger, transport *http.Transport, cfg Config) PolicyTemplates {
	extClient, err := extendedregistry.GetExtendedClient(cfg.Server,
		log, &extendedregistry.Config{
			Address:  "https://" + cfg.Server,
			Username: " ",
			Password: " ",
		},
		transport)
	if err != nil {
		log.Err(err)
	}

	return &oci{
		logger:    log,
		extClient: extClient,
		transport: transport,
		ctx:       ctx,
		cfg:       cfg,
	}
}

// Lists the policy templates
func (o *oci) ListRepos(org, tag string) ([]string, error) {
	var templateRepos []string

	policyRepo, err := o.extClient.ListPublicRepos(org, &api.PaginationRequest{Token: "", Size: -1})
	if err != nil {
		return nil, err
	}

	for _, repo := range policyRepo.Images {
		valid, err := o.extClient.IsValidTag(org, repo.Name, tag)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tags for '%s'", repo.Name)
		}

		if valid {
			templateRepos = append(templateRepos, repo.Name)
		}
	}

	return templateRepos, nil
}

// Loads a policy template into a fs.FS
func (o *oci) Load(userRef string) (fs.FS, error) {
	ref, err := parser.CalculatePolicyRef(userRef, o.cfg.Server)
	if err != nil {
		return nil, err
	}

	descriptorDigest, err := o.pullRef(ref)
	if err != nil {
		return nil, err
	}

	bundleFilePath := filepath.Join(o.cfg.PolicyRoot, "blobs", "sha256", descriptorDigest)

	return loadTarGz(bundleFilePath)
}

func (o *oci) pullRef(ref string) (string, error) {
	ociStore, err := content.NewOCIStore(o.cfg.PolicyRoot)
	if err != nil {
		return "", err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return "", err
	}

	opts := []oras.PullOpt{
		oras.WithContentProvideIngester(ociStore),
	}

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: o.getHosts,
	})

	_, descriptors, err := oras.Pull(o.ctx, resolver, ref, ociStore,
		opts...,
	)

	if err != nil {
		return "", errors.Wrap(err, "oras pull failed")
	}

	if len(descriptors) != 1 {
		return "", errors.Errorf("unexpected layer count of [%d] from the registry; expected 1", len(descriptors))
	}

	ociStore.AddReference(ref, descriptors[0])
	err = ociStore.SaveIndex()
	if err != nil {
		return "", err
	}
	return descriptors[0].Digest.Encoded(), nil
}

func loadTarGz(bundleFilePath string) (fs.FS, error) {
	gzipStream, err := os.Open(bundleFilePath)

	if err != nil {
		return nil, errors.Wrap(err, "failed to open bundle file")
	}
	defer gzipStream.Close()

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gzip reader")
	}

	tfs, err := tarfs.New(uncompressedStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tarfs")
	}

	return tfs, nil
}

func (o *oci) getHosts(server string) ([]docker.RegistryHost, error) {
	client := &http.Client{Transport: o.transport}

	registryHost := docker.RegistryHost{
		Host:         server,
		Scheme:       "https",
		Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
		Client:       client,
		Path:         "/v2",
		Authorizer: docker.NewDockerAuthorizer(
			docker.WithAuthClient(client),
			docker.WithAuthCreds(func(s string) (string, string, error) {
				return " ", " ", nil
			})),
	}

	return []docker.RegistryHost{registryHost}, nil
}
