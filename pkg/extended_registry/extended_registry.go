package extendedregistry

import (
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/aserto-dev/go-utils/fsutil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type Config struct {
	Address        string `json:"address"`
	GRPCAddress    string `json:"extended"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	LocalInfoCache string `json:"local_info_cache"`
}

type ExtendedClient interface {
	ListOrgs(ctx context.Context, page *api.PaginationRequest) (*registry.ListOrgsResponse, error)
	ListRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error)
	ListPublicRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error)
	ListTags(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error)
	ListDigests(ctx context.Context, org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoDigest, *api.PaginationResponse, error)
	GetTag(ctx context.Context, org, repo, tag string) (*api.RegistryRepoTag, error)
	SetVisibility(ctx context.Context, org, repo string, public bool) error
	RemoveImage(ctx context.Context, org, repo, tag string) error
	IsValidTag(ctx context.Context, org, repo, tag string) (bool, error)
	RepoAvailable(ctx context.Context, org, repo string) (bool, error)
	CreateRepo(ctx context.Context, org, repo string) error
}

type xClient struct {
	cfg    *Config
	logger *zerolog.Logger
	client *http.Client
	info   map[string]interface{}
}

func newExtendedClient(logger *zerolog.Logger, cfg *Config, client *http.Client) *xClient {
	return &xClient{
		cfg:    cfg,
		logger: logger,
		client: client,
	}
}

//TODO: This needs to be smarted - rework in progress
func GetExtendedClient(ctx context.Context, server string, logger *zerolog.Logger, cfg *Config, transport *http.Transport) (ExtendedClient, error) {
	httpClient := http.Client{}
	httpClient.Transport = transport

	if server == "ghcr.io" {
		return newGHCRClient(logger,
			&Config{
				Address:  cfg.Address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			&httpClient), nil
	}
	client := newExtendedClient(logger, cfg, &httpClient)
	extendedGRPCAddress, err := client.HasGRPCExtendedAddress()
	if err != nil {
		logger.Debug().Err(err).Str("server", server).Msg("server does not support extended registry")
		return client, nil
	}
	if extendedGRPCAddress != "" {
		asertoClient, err := newAsertoClient(
			ctx,
			logger,
			&Config{
				Address:     cfg.Address,
				GRPCAddress: extendedGRPCAddress,
				Username:    cfg.Username,
				Password:    cfg.Password,
			})
		return asertoClient, err
	}
	return client, errors.Errorf("server does not support extended registry [%s]", server)
}

//TODO: Implement as OCI specific client
func (c *xClient) ListOrgs(ctx context.Context, page *api.PaginationRequest) (*registry.ListOrgsResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *xClient) ListRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) ListPublicRepos(ctx context.Context, org string, page *api.PaginationRequest) (*registry.ListPublicImagesResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *xClient) ListTags(ctx context.Context, org, repo string, page *api.PaginationRequest, deep bool) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) ListDigests(ctx context.Context, org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoDigest, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) GetTag(ctx context.Context, org, repo, tag string) (*api.RegistryRepoTag, error) {
	return nil, errors.New("not implemented")
}

func (c *xClient) SetVisibility(ctx context.Context, org, repo string, public bool) error {
	return errors.New("not implemented")
}

func (c *xClient) RemoveImage(ctx context.Context, org, repo, tag string) error {
	return errors.New("not implemented")
}

func (c *xClient) IsValidTag(ctx context.Context, org, repo, tag string) (bool, error) {
	return false, errors.New("not implemented")
}

func (c *xClient) RepoAvailable(ctx context.Context, org, repo string) (bool, error) {
	return false, errors.New("not implemented")
}

func (c *xClient) CreateRepo(ctx context.Context, org, repo string) error {
	return errors.New("not implemented")
}

func (c *xClient) HasGRPCExtendedAddress() (string, error) {
	var (
		resp            string
		err             error
		infoCacheExists bool
	)

	if c.cfg.LocalInfoCache != "" {
		infoCacheExists, err = fsutil.FileExists(c.cfg.LocalInfoCache)
		if err != nil {
			return "", errors.Wrapf(err, "failed to check if file exists [%s]", c.cfg.LocalInfoCache)
		}
	}

	if infoCacheExists {
		// look inside the local cache to see if we can find our info response for this server
		// if we can, then we can use the grpc address from there
		fileContents, err := ioutil.ReadFile(c.cfg.LocalInfoCache)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read local info cache [%s]", c.cfg.LocalInfoCache)
		}

		resp = string(fileContents)
	} else {

		strURL := c.cfg.Address + "/info"
		resp, err = c.get(strURL)
		if err != nil {
			return "", errors.Wrap(err, "failed to get /info")
		}
	}
	extendedAPIaddress := gjson.Get(resp, "grpc_extended_api").String()
	if extendedAPIaddress == "" {
		return "", errors.New("no extended api endpoint defined in info call")
	}

	err = os.MkdirAll(filepath.Dir(c.cfg.LocalInfoCache), 0700)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create local info cache directory [%s]", c.cfg.LocalInfoCache)
	}
	err = ioutil.WriteFile(c.cfg.LocalInfoCache, []byte(resp), 0600)
	if err != nil {
		return "", errors.Wrap(err, "failed to write info response to local cache")
	}

	return extendedAPIaddress, nil
}

func (c *xClient) get(urlStr string) (string, error) {
	c.logger.Trace().Str("url", urlStr).Msg("extended api get start")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	req := &http.Request{
		URL:    parsedURL,
		Method: "GET",
		Header: http.Header{
			"Authorization": []string{"basic " + base64.URLEncoding.EncodeToString([]byte(c.cfg.Username+":"+c.cfg.Password))},
		},
	}
	response, err := c.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "get failed")
	}

	strBody := &strings.Builder{}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			c.logger.Trace().Err(err).Msg("failed to close response body")
		}

		c.logger.Trace().Str("url", urlStr).Str("body", strBody.String()).Msg("extended api get end")
	}()

	_, err = io.Copy(strBody, response.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("get failed with status code [%d]", response.StatusCode)
	}

	return strBody.String(), nil
}
