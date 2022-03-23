package extendedregistry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	registryClient "github.com/aserto-dev/aserto-go/client/registry"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AsertoClient struct {
	cfg       *Config
	extension *registryClient.Client
}

// TODO Use aserto-go SDK registry client
func NewAsertoClient(logger *zerolog.Logger, cfg *Config) (ExtendedClient, error) {
	var options []client.ConnectionOption
	options = append(options, client.WithAddr(cfg.GRPCAddress),
		client.WithAPIKeyAuth(base64.URLEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))),
		client.WithInsecure(true))
	extensionClient, err := registryClient.New(
		context.Background(),
		options...,
	)
	return &AsertoClient{
		cfg:       cfg,
		extension: extensionClient,
	}, err
}

func (c *AsertoClient) ListOrgs(page *api.PaginationRequest) (*registry.ListOrgsResponse, error) {
	orgs, err := c.extension.Registry.ListOrgs(context.Background(), &registry.ListOrgsRequest{Page: page})
	return orgs, err
}

func (c *AsertoClient) ListRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	// TODO: Asero ListImages does not include pagination and does not allow paginated requests
	resp, err := c.extension.Registry.ListImages(context.Background(), &registry.ListImagesRequest{})
	return resp, nil, err
}

func (c *AsertoClient) ListPublicRepos(org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error) {
	resp, err := c.extension.Registry.ListPublicImages(context.Background(), &registry.ListPublicImagesRequest{Page: page, Organization: org})
	return resp, err
}
func (c *AsertoClient) ListTags(org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	resp, err := c.extension.Registry.ListTagsWithDetails(context.Background(), &registry.ListTagsWithDetailsRequest{
		Page:         page,
		Organization: org,
		Repo:         strings.TrimPrefix(repo, org),
	})
	if !strings.Contains(err.Error(), "unknown method ListTagsWithDetails") {
		if resp != nil {
			return resp.Tag, resp.Page, err
		}
	}
	// Fallback to use remote call if ListTagsWithDetails is unknown
	// Repo name contains the org as org/repo as a response from list repos
	server := strings.TrimPrefix(c.cfg.Address, "https://")
	repoInfo, err := name.NewRepository(fmt.Sprintf("%s/%s", server, repo))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid repo name [%s]", repoInfo)
	}

	// TODO: add paging options
	tags, err := remote.List(repoInfo,
		remote.WithAuth(&authn.Basic{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}),
	)
	if err != nil {
		return nil, nil, err
	}
	var response []*api.RegistryRepoTag
	for i := range tags {
		response = append(response, &api.RegistryRepoTag{Name: tags[i]})
	}

	return response, nil, nil
}

func (c *AsertoClient) GetTag(org, repo, tag string) (*api.RegistryRepoTag, error) {
	image := fmt.Sprintf("%s/%s/%s:%s", strings.TrimPrefix(c.cfg.Address, "https://"), org, repo, tag)
	repoInfo, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	descriptor, err := remote.Get(repoInfo,
		remote.WithAuth(&authn.Basic{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}))

	if err != nil {
		return nil, errors.Wrap(err, "failed to get descriptor")
	}
	var annotations []*api.RegistryRepoAnnotation
	var created time.Time
	var man *v1.Manifest
	err = json.Unmarshal(descriptor.Manifest, &man)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal manifest")
	}
	for i := range man.Layers {
		for k, v := range man.Layers[i].Annotations {
			annotations = append(annotations, &api.RegistryRepoAnnotation{Key: k, Value: v})
		}
	}
	if val, ok := man.Layers[0].Annotations["org.opencontainers.image.created"]; ok {
		created, err = time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse created at time")
		}
	}
	return &api.RegistryRepoTag{
		Name:        descriptor.Ref.Name(),
		Digest:      descriptor.Digest.String(),
		Size:        man.Layers[0].Size,
		Annotations: annotations,
		CreatedAt:   timestamppb.New(created),
	}, nil
}

func (c *AsertoClient) SetVisibility(org, repo string, public bool) error {
	_, err := c.extension.Registry.SetImageVisibility(context.Background(), &registry.SetImageVisibilityRequest{
		Image:        repo,
		Organization: org,
		Public:       public,
	})
	return err
}
func (c *AsertoClient) RemoveImage(org, repo, tag string) error {
	_, err := c.extension.Registry.RemoveImage(context.Background(), &registry.RemoveImageRequest{
		Image:        repo,
		Tag:          tag,
		Organization: org,
	})
	return err
}

func (c *AsertoClient) IsValidTag(org, repo, tag string) (bool, error) {
	_, err := c.GetTag(org, repo, tag)
	if err != nil {
		return false, err
	}
	return true, nil
}
