package extendedregistry

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/aserto-dev/go-utils/cerr"
	"github.com/aserto-dev/scc-lib/sources"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-github/v43/github"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GHCRClient struct {
	base         *xClient
	githubClient *github.Client
	sccClient    sources.Source
}

const public = "public"

var packageType = "container"

// when page size is -1 grab loop through all pages
func newGHCRClient(logger *zerolog.Logger, cfg *Config, client *http.Client) ExtendedClient {
	baseClient := newExtendedClient(logger, cfg, client)
	tp := github.BasicAuthTransport{
		Username:  strings.TrimSpace(cfg.Username),
		Password:  strings.TrimSpace(cfg.Password),
		Transport: client.Transport,
	}

	return &GHCRClient{
		base:         baseClient,
		githubClient: github.NewClient(tp.Client()),
		sccClient:    sources.NewGithub(logger, &sources.Config{CreateRepoTimeoutSeconds: 10}),
	}
}

func (g *GHCRClient) ListOrgs(ctx context.Context, page *api.PaginationRequest) (*registry.ListOrgsResponse, error) {
	organizations, pageInfo, err := g.sccClient.ListOrgs(ctx,
		&sources.AccessToken{Token: g.base.cfg.Password, Type: "Bearer"},
		page)
	if err != nil {
		return nil, errors.Wrap(err, "could not list organizations")
	}
	var response []*api.RegistryOrg
	for i := range organizations {
		response = append(response, &api.RegistryOrg{Name: strings.ToLower(organizations[i].Id)})
	}
	return &registry.ListOrgsResponse{Orgs: response, Page: pageInfo}, nil
}

func (g *GHCRClient) ListRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	g.base.logger.Debug().Msg("List images")
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, nil, err
	}
	paginationResponse := &api.PaginationResponse{}
	var response []*api.PolicyImage
	for {
		resp, pageInfo, err := g.listRepos(ctx, org, nil, github.ListOptions{
			Page:    pageNumber,
			PerPage: pageSize,
		})
		if err != nil {
			return nil, nil, err
		}

		for i := range resp {
			policy := api.PolicyImage{}
			policy.Name = strings.ToLower(*resp[i].Owner.Login) + "/" + strings.ToLower(*resp[i].Name)
			if *resp[i].Visibility == public {
				policy.Public = true
			} else {
				policy.Public = false
			}

			response = append(response, &policy)
		}
		if pageSize != -1 {
			paginationResponse.ResultSize = int32(len(response))
			if pageInfo.NextPage != 0 {
				paginationResponse.NextToken = fmt.Sprintf("%d", pageInfo.NextPage)
			}

			return &registry.ListImagesResponse{Images: response}, paginationResponse, nil
		}
		if pageInfo.NextPage < 1 {
			break
		}
		pageNumber = pageInfo.NextPage
	}

	paginationResponse.ResultSize = int32(len(response))
	return &registry.ListImagesResponse{Images: response}, paginationResponse, nil
}

func (g *GHCRClient) ListPublicRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error) {
	visibility := public
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, err
	}
	resp, pageInfo, err := g.listRepos(ctx, org, &visibility, github.ListOptions{
		Page:    pageNumber,
		PerPage: pageSize,
	})
	if err != nil {
		return nil, err
	}

	var response []*api.PolicyImage
	for i := range resp {
		policy := api.PolicyImage{}
		policy.Name = *resp[i].Owner.Login + "/" + *resp[i].Name
		policy.Public = true

		response = append(response, &policy)
	}

	paginationResponse := &api.PaginationResponse{
		ResultSize: int32(pageInfo.ContentLength),
		TotalSize:  int32(pageInfo.LastPage),
	}
	if pageInfo.NextPage != 0 {
		paginationResponse.NextToken = fmt.Sprintf("%d", pageInfo.NextPage)
	}

	return &registry.ListPublicImagesResponse{
		Images: response,
		Page:   paginationResponse,
	}, nil
}

func (g *GHCRClient) SetVisibility(ctx context.Context, org, repo string, public bool) error {
	return errors.New("not supported. Please set the visibility using the web UI")
}

// ListTags returns tags on the image - if org is empty it returns the tags of the user's image
func (g *GHCRClient) ListTags(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, nil, err
	}

	pieces := strings.Split(repo, "/")
	if len(pieces) == 2 {
		repo = pieces[1]
	}

	tagDetails, pageInfo, err := g.listTagInformation(ctx, org, repo, pageNumber, pageSize)
	if err != nil {
		return nil, nil, err
	}

	var response []*api.RegistryRepoTag
	for i := range tagDetails {
		tagMetadata := tagDetails[i].GetMetadata()
		if tagMetadata == nil || tagMetadata.Container == nil {
			continue
		}
		for _, tag := range tagMetadata.Container.Tags {
			response = append(response, &api.RegistryRepoTag{
				CreatedAt:   timestamppb.New(tagDetails[i].GetCreatedAt().Time),
				Digest:      tagDetails[i].GetName(),
				Name:        tag,
				Size:        0,
				Annotations: nil,
			})
		}

	}

	if pageInfo != nil {
		paginationResponse := &api.PaginationResponse{
			ResultSize: int32(pageInfo.ContentLength),
			TotalSize:  int32(pageInfo.LastPage),
		}
		if pageInfo.NextPage != 0 {
			paginationResponse.NextToken = fmt.Sprintf("%d", pageInfo.NextPage)
		}

		return response, paginationResponse, nil
	}
	return response, nil, nil
}

func (g *GHCRClient) GetTag(ctx context.Context, org, repo, tag string) (*api.RegistryRepoTag, error) {
	tagDetails, _, err := g.listTagInformation(ctx, org, repo, 1, -1) // check all tags
	if err != nil {
		return nil, err
	}

	for i := range tagDetails {
		tagMetadata := tagDetails[i].GetMetadata()
		if tagMetadata == nil || tagMetadata.Container == nil {
			continue
		}
		for _, containerTag := range tagMetadata.Container.Tags {
			if tag == containerTag {
				return &api.RegistryRepoTag{
					CreatedAt:   timestamppb.New(tagDetails[i].GetCreatedAt().Time),
					Digest:      tagDetails[i].GetName(),
					Name:        tag,
					Size:        0,
					Annotations: nil,
				}, nil
			}
		}
	}

	return nil, nil
}

// If tag not specified remove repository
func (g *GHCRClient) RemoveImage(ctx context.Context, org, repo, tag string) error {

	pieces := strings.Split(repo, "/")
	if len(pieces) == 2 {
		repo = pieces[1]
	}

	if tag == "" {
		return g.deletePackage(ctx, org, repo)
	}
	tagDetails, _, err := g.listTagInformation(ctx, org, repo, 1, -1) // check all tags
	if err != nil {
		return err
	}

	for i := range tagDetails {
		containerTags := strings.Join(tagDetails[i].GetMetadata().Container.Tags, ",")
		if strings.Contains(containerTags, tag) {
			return g.deletePackageVersion(ctx, org, repo, tagDetails[i].GetID())
		}
	}

	return nil
}

func (g *GHCRClient) IsValidTag(ctx context.Context, org, repo, tag string) (bool, error) {
	image := fmt.Sprintf("%s/%s/%s:%s", "ghcr.io", org, repo, tag)
	valid, err := g.validImage(ctx, image, g.base.cfg.Username, g.base.cfg.Password)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (g *GHCRClient) CreateRepo(ctx context.Context, org, repo string) error {
	return errors.New("not implemented")
}

func (g *GHCRClient) ListDigests(ctx context.Context, org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoDigest, *api.PaginationResponse, error) {
	tagDetails, _, err := g.listTagInformation(ctx, org, repo, 0, -1)
	if err != nil {
		return nil, nil, err
	}

	pkgVersionsByDigest := make(map[string][]*github.PackageVersion)

	groupPackagesByDigest(pkgVersionsByDigest, tagDetails)

	var result []*api.RegistryRepoDigest

	var digestNames []string

	for digest := range pkgVersionsByDigest {
		digestNames = append(digestNames, digest)
	}

	_, _, digestNamePaged, paginationResponse, err := paginate(
		digestNames,
		func(i, j int) bool {
			if len(pkgVersionsByDigest[digestNames[i]]) == 0 || len(pkgVersionsByDigest[digestNames[j]]) == 0 {
				return false
			}
			return pkgVersionsByDigest[digestNames[i]][0].GetCreatedAt().After(pkgVersionsByDigest[digestNames[j]][0].GetCreatedAt().Time)
		},
		page)
	if err != nil {
		return nil, nil, err
	}

	for _, digestName := range digestNamePaged {
		var tagNames []string

		for _, pkg := range pkgVersionsByDigest[digestName] {
			pkgMetadata := pkg.GetMetadata()
			if pkgMetadata == nil || pkgMetadata.Container == nil {
				continue
			}

			tagNames = append(tagNames, pkgMetadata.Container.Tags...)
		}

		result = append(result, &api.RegistryRepoDigest{
			Digest:    digestName,
			Tags:      tagNames,
			CreatedAt: timestamppb.New(pkgVersionsByDigest[digestName][0].GetCreatedAt().Time),
		})
	}

	return result, paginationResponse, nil
}

func groupPackagesByDigest(packagesByDigest map[string][]*github.PackageVersion, packageVersion []*github.PackageVersion) {
	for _, pv := range packageVersion {
		digest := pv.GetName()
		if _, ok := packagesByDigest[digest]; !ok {
			packagesByDigest[digest] = []*github.PackageVersion{}
		}
		packagesByDigest[digest] = append(packagesByDigest[digest], pv)
	}
}

func (g GHCRClient) deletePackageVersion(ctx context.Context, org, repo string, version int64) error {
	if org == "" || strings.EqualFold(org, g.base.cfg.Username) {
		_, err := g.githubClient.Users.PackageDeleteVersion(ctx, "", packageType, repo, version)
		if err != nil {
			return errors.Wrapf(err, "failed to remove package version %d", version)
		}
		return nil
	}
	_, err := g.githubClient.Organizations.PackageDeleteVersion(ctx, org, packageType, repo, version)
	if err != nil {
		return errors.Wrapf(err, "failed to remove package version %d", version)
	}
	return nil
}

func (g GHCRClient) deletePackage(ctx context.Context, org, repo string) error {
	if org == "" || strings.EqualFold(org, g.base.cfg.Username) {
		_, err := g.githubClient.Users.DeletePackage(ctx, "", packageType, repo)
		if err != nil {
			return errors.Wrap(err, "failed to remove package")
		}
		return nil
	}
	_, err := g.githubClient.Organizations.DeletePackage(ctx, org, packageType, repo)
	if err != nil {
		return errors.Wrap(err, "failed to remove package")
	}

	return nil
}

func (g *GHCRClient) listRepos(ctx context.Context, org string, visibility *string, listOptions github.ListOptions) ([]*github.Package, *github.Response, error) {
	var resp []*github.Package
	var pageInfo *github.Response
	var err error

	if org == "" || strings.EqualFold(org, g.base.cfg.Username) {
		resp, pageInfo, err = g.githubClient.Users.ListPackages(ctx, "",
			&github.PackageListOptions{PackageType: &packageType, Visibility: visibility, ListOptions: listOptions})
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to list container type packages from ghcr")
		}
	} else {
		resp, pageInfo, err = g.githubClient.Organizations.ListPackages(ctx, org,
			&github.PackageListOptions{PackageType: &packageType, Visibility: visibility, ListOptions: listOptions})
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to list containers for org %s", org)
		}
	}
	return resp, pageInfo, nil
}

func (g *GHCRClient) listTagInformation(ctx context.Context, org, repo string, page, size int) ([]*github.PackageVersion, *github.Response, error) {
	var response []*github.PackageVersion
	var err error
	perPage := size
	for {
		// Grab all package versions
		if size == -1 {
			perPage = 100 // max allowed by github api
		}
		var versions []*github.PackageVersion
		var pageInfo *github.Response
		if org == "" || strings.EqualFold(org, g.base.cfg.Username) {
			versions, pageInfo, err = g.githubClient.Users.PackageGetAllVersions(ctx, "", packageType, repo,
				&github.PackageListOptions{
					PackageType: &packageType,
					ListOptions: github.ListOptions{
						Page:    page,
						PerPage: perPage,
					},
				})
			if err != nil {
				ghErr, ok := err.(*github.ErrorResponse)
				if ok && ghErr.Response.StatusCode == http.StatusNotFound {
					return nil, nil, cerr.ErrPolicyNotFound
				}
				return nil, nil, errors.Wrap(err, "failed to list container versions")
			}

		} else {
			versions, pageInfo, err = g.githubClient.Organizations.PackageGetAllVersions(ctx, org, packageType, repo,
				&github.PackageListOptions{
					PackageType: &packageType,
					ListOptions: github.ListOptions{
						Page:    page,
						PerPage: size,
					},
				})
			if err != nil {
				ghErr, ok := err.(*github.ErrorResponse)
				if ok && ghErr.Response.StatusCode == http.StatusNotFound {
					return nil, nil, cerr.ErrPolicyNotFound
				}
				return nil, nil, errors.Wrap(err, "failed to list container versions")
			}
		}
		response = append(response, versions...)
		if size != -1 {
			return response, pageInfo, nil
		}
		if pageInfo.NextPage < 1 {
			break
		}
		page = pageInfo.NextPage
	}
	return response, nil, nil
}

func (g *GHCRClient) validImage(ctx context.Context, repoName, username, password string) (bool, error) {
	repo, err := name.ParseReference(repoName)
	if err != nil {
		g.base.logger.Err(err)
		return false, err
	}
	descriptor, err := remote.Get(repo,
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		}),
		remote.WithContext(ctx))
	if err != nil {
		g.base.logger.Err(err)
		return false, err
	}
	return strings.Contains(string(descriptor.Manifest), "org.openpolicyregistry.type"), nil
}

func parsePaginationRequest(page *api.PaginationRequest) (int, int, error) {
	pageNumber := 0
	pageSize := 30 // Default github page size value

	if page == nil {
		return pageNumber, pageSize, nil
	}
	if page.Token != "" {
		number, err := strconv.ParseInt(page.Token, 10, 32)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "pagination request token must be a number")
		}
		pageNumber = int(number)
	}
	pageSize = int(page.Size)

	return pageNumber, pageSize, nil
}

func (g *GHCRClient) RepoAvailable(ctx context.Context, org, repo string) (bool, error) {
	var resp *github.Response
	var err error
	if org == "" || strings.EqualFold(org, g.base.cfg.Username) {
		_, resp, err = g.githubClient.Users.GetPackage(ctx, "", packageType, repo)
		if err != nil {
			return false, errors.Wrapf(err, "failed to get package %s for user %s", repo, org)
		}
	} else {
		_, resp, err = g.githubClient.Organizations.GetPackage(ctx, org, packageType, repo)
		if err != nil {
			return false, errors.Wrapf(err, "failed to get package %s for org %s", repo, org)
		}
	}

	if resp.StatusCode == 404 {
		return true, nil
	}

	return false, nil
}
