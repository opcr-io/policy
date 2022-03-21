package extendedregistry

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type Config struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ExtendedClient interface {
	ListOrgs(page *api.PaginationRequest) (*registry.ListOrgsResponse, *api.PaginationResponse, error)
	ListRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error)
	ListPublicRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error)
	ListTags(org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoTag, *api.PaginationResponse, error)
	GetTag(org, repo, tag string) (*api.RegistryRepoTag, error)
	SetVisibility(org, repo string, public bool) error
	RemoveImage(org, repo, tag string) error
	IsValidTag(org, repo, tag string) (bool, error)
}

type xClient struct {
	cfg    *Config
	logger *zerolog.Logger
	client *http.Client
}

func newExtendedClient(logger *zerolog.Logger, cfg *Config, client *http.Client) *xClient {
	return &xClient{
		cfg:    cfg,
		logger: logger,
		client: client,
	}
}

//TODO: This needs to be smarted - rework in progress
func GetExtendedClient(server string, logger *zerolog.Logger, cfg *Config, transport *http.Transport) (ExtendedClient, error) {
	httpClient := http.Client{}
	httpClient.Transport = transport

	if server == "ghcr.io" {
		return NewGHCRClient(logger,
			&Config{
				Address:  cfg.Address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			&httpClient), nil
	}
	client := newExtendedClient(logger, cfg, &httpClient)
	isExtendedInfo, err := client.HasExtendedAddress()
	if err != nil {
		return client, errors.Wrapf(err, "server does not support extended registry [%s]", server)
	}
	if isExtendedInfo {
		return NewAsertoClient(logger,
			&Config{
				Address:  cfg.Address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			&httpClient), nil
	}
	return client, errors.Errorf("server does not support extended registry [%s]", server)
}

//TODO: Implement as OCI specific client
func (c *xClient) ListOrgs(page *api.PaginationRequest) (*registry.ListOrgsResponse, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) ListRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) ListPublicRepos(org string, page *api.PaginationRequest) (*registry.ListImagesResponse, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) ListTags(org, repo string, page *api.PaginationRequest) ([]*api.RegistryRepoTag, *api.PaginationResponse, error) {
	return nil, nil, errors.New("not implemented")
}

func (c *xClient) GetTag(org, repo, tag string) (*api.RegistryRepoTag, error) {
	return nil, errors.New("not implemented")
}

func (c *xClient) SetVisibility(org, repo string, public bool) error {
	return errors.New("not implemented")
}

func (c *xClient) RemoveImage(org, repo, tag string) error {
	return errors.New("not implemented")
}

func (c *xClient) IsValidTag(org, repo, tag string) (bool, error) {
	return false, errors.New("not implemented")
}
func (c *xClient) HasExtendedAddress() (bool, error) {
	strURL := c.cfg.Address + "/info"
	resp, err := c.get(strURL)
	if err != nil {
		return false, errors.Wrap(err, "failed to get /info")
	}
	extendedAPIaddress := gjson.Get(resp, "extended_api").String()
	if extendedAPIaddress == "" {
		return false, errors.New("no exteneded api endpoint defined in info call")
	}
	return true, nil
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

func (c *xClient) delete(urlStr string) (string, error) {
	c.logger.Trace().Str("url", urlStr).Msg("extended api delete start")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	req := &http.Request{
		URL:    parsedURL,
		Method: "DELETE",
		Header: http.Header{
			"Authorization": []string{"basic " + base64.URLEncoding.EncodeToString([]byte(c.cfg.Username+":"+c.cfg.Password))},
		},
	}

	response, err := c.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "delete failed")
	}

	strBody := &strings.Builder{}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			c.logger.Trace().Err(err).Msg("failed to close response body")
		}

		c.logger.Trace().Str("url", urlStr).Str("body", strBody.String()).Msg("extended api delete end")
	}()

	_, err = io.Copy(strBody, response.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("delete failed with status code [%d]", response.StatusCode)
	}

	return strBody.String(), nil
}

func (c *xClient) post(urlStr, payload string) (string, error) {
	c.logger.Trace().Str("url", urlStr).Msg("extended api post start")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}

	req := &http.Request{
		URL:    parsedURL,
		Method: "POST",
		Header: http.Header{
			"Authorization": []string{"basic " + base64.URLEncoding.EncodeToString([]byte(c.cfg.Username+":"+c.cfg.Password))},
		},
		Body: ioutil.NopCloser(strings.NewReader(payload)),
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

		c.logger.Trace().Str("url", urlStr).Str("body", strBody.String()).Msg("extended api post end")
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
