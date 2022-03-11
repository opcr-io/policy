package extendedregistry

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Config struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PolicyImage struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type ExtendedClient interface {
	// TODO - add a verify method - GetExtendedClient that will return the client that matches the address?
	ListOrgs() ([]string, error)
	ListRepos(org string) ([]*PolicyImage, error)
	SetVisibility(image string, public bool) error
	RemoveImage(image, tag string) error
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

func GetExtendedClient(server string, logger *zerolog.Logger, cfg *Config, transport *http.Transport) (ExtendedClient, error) {
	httpClient := http.Client{}
	httpClient.Transport = transport
	switch {
	case server == "opcr.io", strings.Contains(server, "aserto"):
		return NewAsertoClient(logger,
			&Config{
				Address:  cfg.Address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			&httpClient), nil
	case server == "ghcr.io":
		return NewGHCRClient(logger,
			&Config{
				Address:  cfg.Address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			&httpClient), nil
	default:
		return newExtendedClient(logger, cfg, &httpClient), errors.Errorf("server does not support extended registry [%s]", server)
	}
}
func (c *xClient) ListOrgs() ([]string, error) {
	return nil, errors.New("not implemented")
}
func (c *xClient) ListRepos(org string) ([]*PolicyImage, error) {
	return nil, errors.New("not implemented")
}
func (c *xClient) SetVisibility(image string, public bool) error {
	return errors.New("not implemented")
}
func (c *xClient) RemoveImage(image, tag string) error {
	return errors.New("not implemented")
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
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
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
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
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

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return "", errors.Errorf("get failed with status code [%d]", response.StatusCode)
	}

	return strBody.String(), nil
}
