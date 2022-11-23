package extendedregistry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/hashicorp/go-multierror"
	"github.com/jhump/protoreflect/grpcreflect"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	AsertoRegistryServiceName = "aserto.registry.v1.Registry"
)

type AsertoClient struct {
	cfg            *Config
	registryClient registry.RegistryClient
	grpcConnection grpc.ClientConnInterface
}

func newAsertoClient(ctx context.Context, cfg *Config) (ExtendedClient, error) {
	var options []client.ConnectionOption
	options = append(options, client.WithAddr(cfg.GRPCAddress))
	if cfg.Username != "" && cfg.Password != "" {
		options = append(options, client.WithAPIKeyAuth(base64.URLEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))))
	} else if cfg.Password != "" {
		options = append(options, client.WithAPIKeyAuth(cfg.Password))
	}

	connection, err := client.NewConnection(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	extensionClient := registry.NewRegistryClient(
		connection.Conn,
	)
	return &AsertoClient{
		cfg:            cfg,
		registryClient: extensionClient,
		grpcConnection: connection.Conn,
	}, err
}

func (c *AsertoClient) ListOrgs(ctx context.Context, page *api.PaginationRequest) (*registry.ListOrgsResponse, error) {
	orgs, err := c.registryClient.ListOrgs(ctx, &registry.ListOrgsRequest{Page: page})
	return orgs, err
}

func (c *AsertoClient) ListRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {

	// TODO: Aserto ListImages does not include pagination and does not allow paginated requests.
	resp, err := c.registryClient.ListImages(ctx, &registry.ListImagesRequest{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list repos")
	}

	var policyImages []*api.PolicyImage

	for _, repo := range resp.Images {
		pieces := strings.Split(repo.Name, "/")
		if len(pieces) != 2 {
			return nil, nil, errors.Errorf("invalid repo name [%s]", repo.Name)
		}

		if pieces[0] != org {
			continue
		}

		policyImages = append(policyImages, repo)
	}

	result := registry.ListImagesResponse{
		Images: policyImages,
	}

	return &result, nil, err
}

func (c *AsertoClient) ListPublicRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error) {
	resp, err := c.registryClient.ListPublicImages(ctx, &registry.ListPublicImagesRequest{Page: page, Organization: org})
	return resp, err
}
func (c *AsertoClient) ListTags(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	repo = strings.TrimPrefix(repo, org+"/")

	listTagsExists, err := c.grpcMethodExists(ctx, "ListTagsWithDetails")
	if err != nil {
		return nil, nil, err
	}

	if listTagsExists {
		listTagsWithDetailsResponse, err := c.registryClient.ListTagsWithDetails(ctx, &registry.ListTagsWithDetailsRequest{
			Page:         page,
			Organization: org,
			Repo:         repo,
		})
		if err != nil {
			return nil, nil, err
		}

		return listTagsWithDetailsResponse.Tag, listTagsWithDetailsResponse.Page, nil
	}

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
	var annotationsMap map[string]string
	var created time.Time
	var man *v1.Manifest
	err = json.Unmarshal(descriptor.Manifest, &man)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal manifest")
	}

	annotationsMap = man.Annotations
	if annotationsMap == nil {
		annotationsMap = make(map[string]string)
	}

	for i := range man.Layers {
		for k, v := range man.Layers[i].Annotations {
			if i == 0 && k == AnnotationImageCreated {
				created, err = time.Parse(time.RFC3339, v)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse created at time")
				}
			}
			annotationsMap[k] = v
		}
	}

	var size int64
	if len(man.Layers) > 0 {
		size = man.Layers[0].Size
	}

	var annotations []*api.RegistryRepoAnnotation

	for k, v := range annotationsMap {
		annotations = append(annotations, &api.RegistryRepoAnnotation{Key: k, Value: v})
	}

	return &api.RegistryRepoTag{
		Name:        descriptor.Ref.Name(),
		Digest:      descriptor.Digest.String(),
		Size:        size,
		Annotations: annotations,
		CreatedAt:   timestamppb.New(created),
	}, nil
}

func (c *AsertoClient) SetVisibility(ctx context.Context, org, repo string, public bool) error {
	_, err := c.registryClient.SetImageVisibility(ctx, &registry.SetImageVisibilityRequest{
		Image:        repo,
		Organization: org,
		Public:       public,
	})
	return err
}
func (c *AsertoClient) RemoveImage(ctx context.Context, org, repo, tag string) error {
	_, err := c.registryClient.RemoveImage(ctx, &registry.RemoveImageRequest{
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
	repoAvailableResponse, err := c.registryClient.RepoAvailable(ctx, &registry.RepoAvailableRequest{
		Organization: org,
		Repo:         repo,
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if repo [%s] exists", repo)
	}

	if repoAvailableResponse.Availability == api.NameAvailability_NAME_AVAILABILITY_AVAILABLE {
		return true, nil
	}

	return false, nil
}

func (c *AsertoClient) CreateRepo(ctx context.Context, org, repo string) error {
	_, err := c.registryClient.CreateImage(ctx, &registry.CreateImageRequest{
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

func (c *AsertoClient) ListDigests(ctx context.Context, org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoDigest, *api.PaginationResponse, error) {
	listDigestExists, err := c.grpcMethodExists(ctx, "ListDigests")
	if err != nil {
		return nil, nil, err
	}

	if listDigestExists {
		listDigestResponse, err := c.registryClient.ListDigests(ctx, &registry.ListDigestsRequest{
			Page:         page,
			Organization: org,
			Repo:         repo,
		})
		if err != nil {
			return nil, nil, err
		}

		return listDigestResponse.Digests, listDigestResponse.Page, nil
	}

	return c.listDigestsRemote(ctx, org, repo, page)
}

func (c *AsertoClient) listDigestsRemote(ctx context.Context, org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoDigest, *api.PaginationResponse, error) {
	var paginationResponse *api.PaginationResponse

	digestGroups := make(map[string][]*api.RegistryRepoTag)

	listTagsPage := &api.PaginationRequest{
		Size: -1,
	}

	tags, _, err := c.ListTags(ctx, org, repo, listTagsPage, true)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list tags")
	}

	groupByDigest(digestGroups, tags)

	var result []*api.RegistryRepoDigest

	var digestNames []string

	for digest := range digestGroups {
		digestNames = append(digestNames, digest)
	}

	_, _, digestNamePaged, paginationResponse, err := paginate(
		digestNames,
		func(i, j int) bool {
			if len(digestGroups[digestNames[i]]) == 0 || len(digestGroups[digestNames[j]]) == 0 {
				return false
			}
			return digestGroups[digestNames[i]][0].CreatedAt.AsTime().After(digestGroups[digestNames[j]][0].CreatedAt.AsTime())
		},
		page)
	if err != nil {
		return nil, nil, err
	}

	for _, digestName := range digestNamePaged {
		var tagNames []string

		for _, tag := range digestGroups[digestName] {
			tagNames = append(tagNames, tag.Name)
		}

		result = append(result, &api.RegistryRepoDigest{
			Digest:    digestName,
			Tags:      tagNames,
			CreatedAt: digestGroups[digestName][0].CreatedAt,
		})
	}

	return result, paginationResponse, nil
}

func groupByDigest(tagsByDigest map[string][]*api.RegistryRepoTag, tags []*api.RegistryRepoTag) {
	for _, tag := range tags {
		digest := tag.Digest
		if _, ok := tagsByDigest[digest]; !ok {
			tagsByDigest[digest] = []*api.RegistryRepoTag{}
		}
		tagsByDigest[digest] = append(tagsByDigest[digest], tag)
	}
}

func (c *AsertoClient) listTagsRemote(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	server := strings.TrimPrefix(c.cfg.Address, "https://")
	repoName, err := name.NewRepository(server + "/" + org + "/" + repo)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid repo name [%s]", repo)
	}

	tags, err := remote.List(repoName,
		remote.WithAuth(&authn.Basic{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}),
		remote.WithContext(ctx))

	if err != nil {
		return c.handleTransportError(err)
	}

	if len(tags) == 0 {
		return []*api.RegistryRepoTag{}, nil, nil
	}

	start, end, _, nextPage, err := paginate(
		tags,
		func(i, j int) bool {
			return tags[i] > tags[j]
		},
		page)
	if err != nil {
		return nil, nil, err
	}

	ref := server + "/" + org + "/" + repo
	result, err := c.processTags(ctx, tags, ref, start, end, deep)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list tags from registry")
	}

	return result, nextPage, nil
}

func paginate(collection []string, less func(i, j int) bool, page *api.PaginationRequest) (int, int, []string, *api.PaginationResponse, error) {
	sort.Slice(collection, less)

	start := 0
	end := len(collection)

	if page != nil {

		if page.Token != "" {
			pageTokenExists := false
			for i, tag := range collection {
				if tag == page.Token {
					start = i
					pageTokenExists = true
					break
				}
			}
			if !pageTokenExists {
				return 0, 0, nil, nil, errors.Errorf("invalid page token: '%s'", page.Token)
			}
		}

		count := int(page.Size)
		if count > 0 && (start+count) < len(collection) {
			end = start + count
		}
	}

	paginationResponse := &api.PaginationResponse{}
	if end < len(collection) {
		paginationResponse.NextToken = collection[end]
	}
	paginationResponse.ResultSize = int32(end - start)

	return start, end, collection[start:end], paginationResponse, nil
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

func (c *AsertoClient) handleTransportError(err error) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
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

func (c *AsertoClient) grpcMethodExists(ctx context.Context, method string) (bool, error) {
	grpcReflectClient := grpcreflect.NewClientV1Alpha(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(c.grpcConnection))
	defer grpcReflectClient.Reset()

	descriptor, err := grpcReflectClient.ResolveService(AsertoRegistryServiceName)
	if err != nil {
		return false, errors.Wrap(err, "failed to resolve registry service")
	}

	methodDescriptor := descriptor.FindMethodByName(method)

	if methodDescriptor != nil {
		return true, nil
	}

	return false, nil
}
