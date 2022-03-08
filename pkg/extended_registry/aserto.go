package extendedregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AsertoClient struct {
	base *xClient
}

// TODO Use aserto-go SDK registry client
func NewAsertoClient(logger *zerolog.Logger, cfg *Config, client *http.Client) ExtendedClient {
	baseClient := newExtendedClient(logger, cfg, client)

	return &AsertoClient{
		base: baseClient,
	}
}

func (c *AsertoClient) ListOrgs(page *api.PaginationRequest) (*registry.ListOrgsResponse, *api.PaginationResponse, error) {
	address, err := c.extendedAPIAddress()
	if err != nil {
		return nil, nil, err
	}

	jsonBody, err := c.base.get(address + "/api/v1/registry/organizations")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list organizations")
	}
	type Org struct {
		Name string `json:"name"`
	}
	// TODO Add paginated information to read all orgs
	var parseStruct struct {
		Orgs []Org `json:"orgs"`
	}

	err = json.Unmarshal([]byte(jsonBody), &parseStruct)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal orgs")
	}
	var response []*api.RegistryOrg
	for i := range parseStruct.Orgs {
		response = append(response, &api.RegistryOrg{Name: parseStruct.Orgs[i].Name})
	}
	return &registry.ListOrgsResponse{Orgs: response}, nil, nil
}

func (c *AsertoClient) ListRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	if org == "" {
		return nil, nil, nil
	}
	address, err := c.extendedAPIAddress()
	if err != nil {
		return nil, nil, err
	}

	jsonBody, err := c.base.get(address + "/api/v1/registry/images")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list images")
	}

	response := struct {
		Images []*api.PolicyImage `json:"images"`
	}{}

	err = json.Unmarshal([]byte(jsonBody), &response)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal policy list response")
	}

	return &registry.ListImagesResponse{Images: response.Images}, nil, nil
}

func (c *AsertoClient) ListPublicRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	address, err := c.extendedAPIAddress()
	if err != nil {
		return nil, nil, err
	}

	jsonBody, err := c.base.get(fmt.Sprintf("%s/api/v1/registry/images/%s/public?size=%d&token=%s", address, org, page.Size, page.Token))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list public images")
	}

	response := struct {
		Page   api.PaginationResponse `json:"page"`
		Images []*api.PolicyImage     `json:"images"`
	}{}

	err = json.Unmarshal([]byte(jsonBody), &response)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal policy list response")
	}
	fmt.Println(jsonBody)

	return &registry.ListImagesResponse{Images: response.Images}, &response.Page, nil
}
func (c *AsertoClient) ListTags(org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	// Repo name contains the org as org/repo as a response from list repos
	repoInfo, err := name.NewRepository(fmt.Sprintf("%s/%s", strings.TrimPrefix(c.base.cfg.Address, "https://"), repo))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid repo name [%s]", repoInfo)
	}

	// TODO: add paging options
	tags, err := remote.List(repoInfo,
		remote.WithAuth(&authn.Basic{
			Username: c.base.cfg.Username,
			Password: c.base.cfg.Password,
		}))
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
	image := fmt.Sprintf("%s/%s/%s:%s", strings.TrimPrefix(c.base.cfg.Address, "https://"), org, repo, tag)
	repoInfo, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	descriptor, err := remote.Get(repoInfo,
		remote.WithAuth(&authn.Basic{
			Username: c.base.cfg.Username,
			Password: c.base.cfg.Password,
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
	image := fmt.Sprintf("%s/%s", org, repo)
	address, err := c.extendedAPIAddress()
	if err != nil {
		return err
	}

	toUpdate := address + "/api/v1/registry/images/" + image + "/visibility"

	// TODO: error handling from body/header
	_, err = c.base.post(toUpdate, fmt.Sprintf(`{"public": %t}`, public))
	if err != nil {
		return errors.Wrap(err, "failed to update image visibility")
	}

	return nil
}
func (c *AsertoClient) RemoveImage(org, repo, tag string) error {
	image := fmt.Sprintf("%s/%s", org, repo)
	address, err := c.extendedAPIAddress()
	if err != nil {
		return err
	}

	toDelete := address + "/api/v1/registry/images/" + image
	if tag != "" {
		toDelete += "?tag=" + url.QueryEscape(tag)
	}

	// TODO: error handling from body/header
	_, err = c.base.delete(toDelete)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}

	return nil
}

func (c *AsertoClient) IsValidTag(org, repo, tag string) (bool, error) {
	_, err := c.GetTag(org, repo, tag)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *AsertoClient) extendedAPIAddress() (string, error) {
	strURL := c.base.cfg.Address + "/info"
	response, err := c.base.get(strURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to get /info")
	}

	return "https://" + gjson.Get(response, "extended_api").String(), nil
}
