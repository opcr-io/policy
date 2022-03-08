package extendedregistry

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
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

func NewGHCRClient(logger *zerolog.Logger, cfg *Config, client *http.Client) ExtendedClient {
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

func (g *GHCRClient) ListOrgs(page *api.PaginationRequest) (*registry.ListOrgsResponse, *api.PaginationResponse, error) {
	orgs, pageInfo, err := g.sccClient.ListOrgs(context.Background(),
		&sources.AccessToken{Token: g.base.cfg.Password, Type: "Bearer"},
		page)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not list organizations")
	}
	var response []*api.RegistryOrg
	for i := range orgs {
		response = append(response, &api.RegistryOrg{Name: orgs[i]})
	}
	return &registry.ListOrgsResponse{Orgs: response}, pageInfo, nil
}

func (g *GHCRClient) ListRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	g.base.logger.Debug().Msg("List images")
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, nil, err
	}
	resp, _, err := g.listRepos(org, nil, github.ListOptions{
		Page:    pageNumber,
		PerPage: pageSize,
	})
	if err != nil {
		return nil, nil, err
	}

	var response []*api.PolicyImage
	for i := range resp {
		policy := api.PolicyImage{}
		policy.Name = *resp[i].Name
		if *resp[i].Visibility == public {
			policy.Public = true
		} else {
			policy.Public = false
		}

		response = append(response, &policy)
	}

	return &registry.ListImagesResponse{Images: response}, nil, nil
}

func (g *GHCRClient) ListPublicRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	visibility := public
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, nil, err
	}
	resp, _, err := g.listRepos(org, &visibility, github.ListOptions{
		Page:    pageNumber,
		PerPage: pageSize,
	})
	if err != nil {
		return nil, nil, err
	}

	var response []*api.PolicyImage
	for i := range resp {
		policy := api.PolicyImage{}
		policy.Name = *resp[i].Owner.Login + "/" + *resp[i].Name
		policy.Public = true

		response = append(response, &policy)
	}
	return &registry.ListImagesResponse{Images: response}, nil, nil
}

func (g *GHCRClient) SetVisibility(org, repo string, public bool) error {
	return errors.New("not supported. Please set the visibility using the web UI")
}

// ListTags returns tags on the image - if org is empty it returns the tags of the user's image
func (g *GHCRClient) ListTags(org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	pageNumber, pageSize, err := parsePaginationRequest(page)
	if err != nil {
		return nil, nil, err
	}
	tagDetails, err := g.listTagInformation(org, repo, pageNumber, pageSize)
	if err != nil {
		return nil, nil, err
	}

	var response []*api.RegistryRepoTag
	for i := range tagDetails {
		response = append(response, &api.RegistryRepoTag{
			CreatedAt:   timestamppb.New(tagDetails[i].GetCreatedAt().Time),
			Digest:      tagDetails[i].GetName(),
			Name:        strings.Join(tagDetails[i].GetMetadata().Container.Tags, ","),
			Size:        0,
			Annotations: nil,
		})
	}
	return response, nil, nil
}

func (g *GHCRClient) GetTag(org, repo, tag string) (*api.RegistryRepoTag, error) {
	tagDetails, err := g.listTagInformation(org, repo, 1, 100) // check the latest 100 tags
	if err != nil {
		return nil, err
	}

	for i := range tagDetails {
		containerTags := strings.Join(tagDetails[i].GetMetadata().Container.Tags, ",")
		if strings.Contains(containerTags, tag) {
			return &api.RegistryRepoTag{
				CreatedAt:   timestamppb.New(tagDetails[i].GetCreatedAt().Time),
				Digest:      tagDetails[i].GetName(),
				Name:        containerTags,
				Size:        0,
				Annotations: nil,
			}, nil
		}
	}

	return nil, nil
}

// If tag not specified remove repository
func (g *GHCRClient) RemoveImage(org, repo, tag string) error {
	if tag == "" {
		return g.deletePackage(org, repo)
	}
	tagDetails, err := g.listTagInformation(org, repo, 1, 100) // check the latest 100 tags
	if err != nil {
		return err
	}

	for i := range tagDetails {
		containerTags := strings.Join(tagDetails[i].GetMetadata().Container.Tags, ",")
		if strings.Contains(containerTags, tag) {
			return g.deletePackageVersion(org, repo, tagDetails[i].GetID())
		}
	}

	return nil
}

func (g *GHCRClient) IsValidTag(org, repo, tag string) (bool, error) {
	image := fmt.Sprintf("%s/%s/%s:%s", "ghcr.io", org, repo, tag)
	valid, err := g.validImage(image, g.base.cfg.Username, g.base.cfg.Password)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (g GHCRClient) deletePackageVersion(org, repo string, version int64) error {
	if org == "" || org == g.base.cfg.Username {
		_, err := g.githubClient.Users.PackageDeleteVersion(context.Background(), "", packageType, repo, version)
		if err != nil {
			return errors.Wrapf(err, "failed to remove package version %d", version)
		}
		return nil
	}
	_, err := g.githubClient.Organizations.PackageDeleteVersion(context.Background(), org, packageType, repo, version)
	if err != nil {
		return errors.Wrapf(err, "failed to remove package version %d", version)
	}
	return nil
}

func (g GHCRClient) deletePackage(org, repo string) error {
	if org == "" || org == g.base.cfg.Username {
		_, err := g.githubClient.Users.DeletePackage(context.Background(), "", packageType, repo)
		if err != nil {
			return errors.Wrap(err, "failed to remove package")
		}
		return nil
	}
	_, err := g.githubClient.Organizations.DeletePackage(context.Background(), org, packageType, repo)
	if err != nil {
		return errors.Wrap(err, "failed to remove package")
	}

	return nil
}

func (g *GHCRClient) listRepos(org string, visibility *string, listOptions github.ListOptions) ([]*github.Package, *github.Response, error) {
	var resp []*github.Package
	var pageInfo *github.Response
	var err error

	if org == "" || org == g.base.cfg.Username {
		resp, pageInfo, err = g.githubClient.Users.ListPackages(context.Background(), "",
			&github.PackageListOptions{PackageType: &packageType, Visibility: visibility, ListOptions: listOptions})
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to list container type packages from ghcr")
		}
	} else {
		resp, pageInfo, err = g.githubClient.Organizations.ListPackages(context.Background(), org,
			&github.PackageListOptions{PackageType: &packageType, Visibility: visibility, ListOptions: listOptions})
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to list containers for org %s", org)
		}
	}
	return resp, pageInfo, nil
}

func (g *GHCRClient) listTagInformation(org, repo string, page, size int) ([]*github.PackageVersion, error) {
	var response []*github.PackageVersion
	var err error
	if org == "" || org == g.base.cfg.Username {
		response, _, err = g.githubClient.Users.PackageGetAllVersions(context.Background(), "", packageType, repo,
			&github.PackageListOptions{
				PackageType: &packageType,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: size,
				},
			})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list container versions")
		}
	} else {
		response, _, err = g.githubClient.Organizations.PackageGetAllVersions(context.Background(), org, packageType, repo,
			&github.PackageListOptions{
				PackageType: &packageType,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: size,
				},
			})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list container versions")
		}
	}
	return response, nil
}

func (g *GHCRClient) validImage(repoName, username, password string) (bool, error) {
	repo, err := name.ParseReference(repoName)
	if err != nil {
		g.base.logger.Err(err)
		return false, err
	}
	descriptor, err := remote.Get(repo,
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		}))
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
	if page.Size > 0 {
		pageSize = int(page.Size)
	}

	return pageNumber, pageSize, nil
}
