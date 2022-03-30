package extendedregistry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	registryClient "github.com/aserto-dev/aserto-go/client/registry"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/hashicorp/go-multierror"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AsertoClient struct {
	cfg       *Config
	extension *registryClient.Client
}

func NewAsertoClient(ctx context.Context, logger *zerolog.Logger, cfg *Config) (ExtendedClient, error) {
	var options []client.ConnectionOption
	options = append(options, client.WithAddr(cfg.GRPCAddress))
	if cfg.Username != "" && cfg.Password != "" {
		options = append(options, client.WithAPIKeyAuth(base64.URLEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))))
	} else if cfg.Password != "" {
		options = append(options, client.WithAPIKeyAuth(cfg.Password))
	}
	extensionClient, err := registryClient.New(
		ctx,
		options...,
	)
	return &AsertoClient{
		cfg:       cfg,
		extension: extensionClient,
	}, err
}

func (c *AsertoClient) ListOrgs(ctx context.Context, page *api.PaginationRequest) (*registry.ListOrgsResponse, error) {
	orgs, err := c.extension.Registry.ListOrgs(ctx, &registry.ListOrgsRequest{Page: page})
	return orgs, err
}

func (c *AsertoClient) ListRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	// TODO: Aserto ListImages does not include pagination and does not allow paginated requests
	resp, err := c.extension.Registry.ListImages(ctx, &registry.ListImagesRequest{})
	return resp, nil, err
}

func (c *AsertoClient) ListPublicRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error) {
	resp, err := c.extension.Registry.ListPublicImages(ctx, &registry.ListPublicImagesRequest{Page: page, Organization: org})
	return resp, err
}
func (c *AsertoClient) ListTags(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	repo = strings.TrimPrefix(repo, org+"/")
	if !deep {
		resp, err := c.extension.Registry.ListTagsWithDetails(ctx, &registry.ListTagsWithDetailsRequest{
			Page:         page,
			Organization: org,
			Repo:         repo,
		})
		if err != nil {
			if !strings.Contains(err.Error(), "unknown method ListTagsWithDetails") {
				return nil, nil, err
			}
		}
		if resp != nil {
			return resp.Tag, resp.Page, err
		}
	}
	// Fallback to use remote call if ListTagsWithDetails is unknown or deep is true
	// Repo name contains the org as org/repo as a response from list repos
	return c.listTagsRemote(ctx, org, repo, page, deep)
}

func (c *AsertoClient) GetTag(ctx context.Context, org, repo, tag string) (*api.RegistryRepoTag, error) {
	image := fmt.Sprintf("%s/%s/%s:%s", strings.TrimPrefix(c.cfg.Address, "https://"), org, repo, tag)
	repoInfo, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	descriptor, err := remote.Get(repoInfo,
		remote.WithAuth(&authn.Basic{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}),
		remote.WithContext(ctx),
	)

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

func (c *AsertoClient) SetVisibility(ctx context.Context, org, repo string, public bool) error {
	_, err := c.extension.Registry.SetImageVisibility(ctx, &registry.SetImageVisibilityRequest{
		Image:        repo,
		Organization: org,
		Public:       public,
	})
	return err
}
func (c *AsertoClient) RemoveImage(ctx context.Context, org, repo, tag string) error {
	_, err := c.extension.Registry.RemoveImage(ctx, &registry.RemoveImageRequest{
		Image:        repo,
		Tag:          tag,
		Organization: org,
	})
	return err
}

func (c *AsertoClient) IsValidTag(ctx context.Context, org, repo, tag string) (bool, error) {
	_, err := c.GetTag(ctx, org, repo, tag)

	tErr, ok := errors.Cause(err).(*transport.Error)
	if ok {
		if tErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *AsertoClient) RepoAvailable(ctx context.Context, org, repo string) (bool, error) {
	repoAvailableResponse, err := c.extension.Registry.RepoAvailable(ctx, &registry.RepoAvailableRequest{
		Organization: org,
		Repo:         repo,
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if repo [%s] exists", repo)
	}

	if repoAvailableResponse.Availability == api.NameAvailability_NAME_AVAILABILITY_UNAVAILABLE {
		return true, nil
	}

	return false, nil
}

func (c *AsertoClient) CreateRepo(ctx context.Context, org, repo string) error {
	_, err := c.extension.Registry.CreateImage(ctx, &registry.CreateImageRequest{
		Organization: org,
		Image: &api.PolicyImage{
			Name: repo,
		},
	})

	if err != nil {
		return errors.Wrap(err, "failed to create repo")
	}

	return nil
}

func (c *AsertoClient) listTagsRemote(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	server := strings.TrimPrefix(c.cfg.Address, "https://")
	repoName, err := name.NewRepository(server + "/" + org + "/" + repo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid repo name [%s]", repo)
	}

	count := 30
	if page != nil {
		count = int(page.Size)
	}

	tags, err := remote.List(repoName,
		remote.WithPageSize(count),
		remote.WithAuth(&authn.Basic{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}),
		remote.WithContext(ctx))

	if err != nil {
		if tErr, ok := err.(*transport.Error); ok {
			switch tErr.StatusCode {
			case http.StatusUnauthorized:
				return nil, nil, errors.Wrap(err, "authentication to registry failed")
			case http.StatusNotFound:
				return []*api.RegistryRepoTag{}, nil, nil
			}
		}

		return nil, nil, errors.Wrap(err, "failed to list tags from registry")
	}

	p := 0
	if page != nil && page.Token != "" {
		p, err = strconv.Atoi(page.Token)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse token as a page number")
		}
	}

	start := p * count
	end := start + count
	if start >= end && count != -1 {
		return []*api.RegistryRepoTag{}, nil, nil
	}

	if end > len(tags) {
		end = len(tags)
	}

	ref := server + "/" + org + "/" + repo
	result, err := c.processTags(ctx, tags, ref, start, end, deep)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list tags from registry")
	}

	nextPage := &api.PaginationResponse{}
	if end+1 < len(tags) {
		nextPage.NextToken = strconv.Itoa(p + 1)
	}

	return result, nextPage, nil
}

func (c *AsertoClient) processTags(ctx context.Context, tags []string, repo string, start, end int, deep bool) ([]*api.RegistryRepoTag, error) {

	wg := &sync.WaitGroup{}
	wg.Add(end - start)

	me := multierror.Error{}
	result := make([]*api.RegistryRepoTag, end-start)

	for i, tag := range tags[start:end] {
		if !deep {
			result[i] = &api.RegistryRepoTag{
				Name: tag,
			}
			wg.Done()
			continue
		}

		go func(i int, tag string) {
			defer wg.Done()

			ref := repo + ":" + tag

			parsedRef, err := name.ParseReference(ref)
			if err != nil {
				me.Errors = append(me.Errors, errors.Wrapf(err, "failed to parse reference [%s]", ref))
			}

			desc, err := remote.Get(parsedRef,
				remote.WithAuth(&authn.Basic{
					Username: c.cfg.Username,
					Password: c.cfg.Password,
				}),
				remote.WithContext(ctx),
			)
			if err != nil {
				me.Errors = append(me.Errors, errors.Wrapf(err, "failed to get image [%s]", ref))
				return
			}

			manifestReader := bytes.NewReader(desc.Manifest)
			m, err := v1.ParseManifest(manifestReader)
			if err != nil {
				me.Errors = append(me.Errors, errors.Wrapf(err, "failed to parse manifest [%s]", ref))
				return
			}
			if len(m.Layers) == 0 {
				me.Errors = append(me.Errors, errors.Errorf("no layers found in manifest [%s]", ref))
				return
			}

			createdAt := time.Unix(0, 0)
			createdAtStr := m.Layers[0].Annotations[ocispec.AnnotationCreated]
			if createdAtStr != "" {
				createdAt, err = time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					me.Errors = append(me.Errors, errors.Errorf("failed to parse createdAt timestamp annotation [%s]", ref))
				}
			}

			var annotations []*api.RegistryRepoAnnotation
			for i := range m.Layers {
				for k, v := range m.Layers[i].Annotations {
					annotations = append(annotations, &api.RegistryRepoAnnotation{Key: k, Value: v})
				}
			}

			result[i] = &api.RegistryRepoTag{
				Name:        tag,
				Digest:      desc.Digest.String(),
				Annotations: annotations,
				Size:        m.Layers[0].Size,
				CreatedAt:   timestamppb.New(createdAt),
			}

		}(i, tag)
	}

	wg.Wait()

	err := me.ErrorOrNil()
	if err != nil {
		return nil, err
	}
	return result, nil
}
