package policytemplates

import (
	"compress/gzip"
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	tarfs "github.com/nlepage/go-tarfs"

	"github.com/containerd/containerd/remotes/docker"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	ociclient "github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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

const policy = "aserto-templates"
const ci = "ci-"

// NewOCI returns a new policy template provider for OCI.
func NewOCI(ctx context.Context, log *zerolog.Logger, httpTransport *http.Transport, cfg Config) PolicyTemplates {
	extClient, err := extendedregistry.GetExtendedClient(
		ctx,
		cfg.Server,
		//TODO: fix extended registry for ghcr to allow both with and without credentials for public images
		log, &extendedregistry.Config{
			Address:  "https://" + cfg.Server,
			Username: os.Getenv("USER"),
			Password: os.Getenv("GIT_TOKEN"),
		},
		httpTransport)
	if err != nil {
		log.Err(err)
	}

	return &oci{
		logger:    log,
		extClient: extClient,
		transport: httpTransport,
		ctx:       ctx,
		cfg:       cfg,
	}
}

// Lists the policy templates.
func (o *oci) ListRepos(org, tag string) (map[string]*api.RegistryRepoTag, error) {
	templateRepos := make(map[string]*api.RegistryRepoTag)

	policyRepo, err := o.extClient.ListPublicRepos(o.ctx, org, &api.PaginationRequest{Token: "", Size: -1})
	if err != nil {
		return nil, err
	}

	for _, repo := range policyRepo.Images {
		// userRef := fmt.Sprintf("%s:%s", repo.Name, tag)
		if strings.Contains(repo.Name, org) {
			repo.Name = strings.TrimPrefix(repo.Name, org+"/")
		}
		apiTag, err := o.extClient.GetTag(o.ctx, org, repo.Name, tag)

		// ref, err := parser.CalculatePolicyRef(userRef, o.cfg.Server)
		// if err != nil {
		// 	return nil, err
		// }

		// ociClient, err := ociclient.NewOCI(o.ctx, o.logger, o.getHosts, o.cfg.PolicyRoot)
		// if err != nil {
		// 	return nil, err
		// }

		// Even pulling the images from aserto-templates the annotation do not contain the description and kind desired
		// _, err = ociClient.Pull(ref)
		// if err != nil {
		// 	return nil, err
		// }
		// descriptos, err := ociClient.ListReferences()
		// if err != nil {
		// 	return nil, err
		// }

		// for annotationKey, annotationValue := range descriptos[ref].Annotations {
		// 	apiTag.Annotations = append(apiTag.Annotations, &api.RegistryRepoAnnotation{
		// 		Key:   annotationKey,
		// 		Value: annotationValue,
		// 	})
		// }
		if strings.Contains(repo.Name, ci) {
			apiTag.Annotations = append(apiTag.Annotations, &api.RegistryRepoAnnotation{Key: extendedregistry.AnnotationPolicyRegistryTemplateKind, Value: "ci-template"})
			apiTag.Annotations = append(apiTag.Annotations, &api.RegistryRepoAnnotation{Key: extendedregistry.AnnotationImageDescription, Value: "CI Templates"})
		}
		if org == policy {
			apiTag.Annotations = append(apiTag.Annotations, &api.RegistryRepoAnnotation{Key: extendedregistry.AnnotationPolicyRegistryTemplateKind, Value: "policy"})
			apiTag.Annotations = append(apiTag.Annotations, &api.RegistryRepoAnnotation{Key: extendedregistry.AnnotationImageDescription, Value: "Policy example templates"})
		}
		tErr, ok := errors.Cause(err).(*transport.Error)
		if ok {
			if tErr.StatusCode == http.StatusNotFound {
				continue
			}
		}

		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tags for '%s'", repo.Name)
		}
		templateRepos[repo.Name] = apiTag

	}

	return templateRepos, nil
}

// Loads a policy template into a fs.FS.
func (o *oci) Load(userRef string) (fs.FS, error) {
	ref, err := parser.CalculatePolicyRef(userRef, o.cfg.Server)
	if err != nil {
		return nil, err
	}

	ociClient, err := ociclient.NewOCI(o.ctx, o.logger, o.getHosts, o.cfg.PolicyRoot)
	if err != nil {
		return nil, err
	}

	digest, err := ociClient.Pull(ref)
	if err != nil {
		return nil, err
	}

	bundleFilePath := filepath.Join(o.cfg.PolicyRoot, "blobs", "sha256", digest.Encoded())

	return loadTarGz(bundleFilePath)
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
		Authorizer:   docker.NewDockerAuthorizer(o.getDockerAuthorizerOptions(client)...),
	}

	return []docker.RegistryHost{registryHost}, nil
}

func (o *oci) getDockerAuthorizerOptions(client *http.Client) []docker.AuthorizerOpt {
	var opts []docker.AuthorizerOpt
	opts = append(opts, docker.WithAuthClient(client))
	//TODO: replace if with credentials configuration
	if o.cfg.Server == "opcr.io" {
		opts = append(opts, docker.WithAuthCreds(func(s string) (string, string, error) {
			return " ", " ", nil
		}))
	}
	return opts
}
