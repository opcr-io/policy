package extendedclient

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type Config struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ExtendedClient struct {
	cfg       *Config
	logger    *zerolog.Logger
	transport *http.Transport
}

func NewExtendedClient(logger *zerolog.Logger, cfg *Config, transport *http.Transport) *ExtendedClient {
	return &ExtendedClient{
		cfg:       cfg,
		logger:    logger,
		transport: transport,
	}
}

func (c *ExtendedClient) Supported() bool {
	address, err := c.extendedAPIAddress()
	if err != nil {
		c.logger.Trace().Err(err).Msg("failed to get extended API address")
		return false
	}

	return address != ""
}

func (c *ExtendedClient) ListImages() ([]string, error) {
	address, err := c.extendedAPIAddress()
	if err != nil {
		return nil, err
	}

	jsonBody, err := c.get(address + "/api/v1/registry/images")
	if err != nil {
		return nil, errors.Wrap(err, "failed to list images")
	}

	images := gjson.Get(jsonBody, "images").Array()

	result := make([]string, len(images))
	for idx, image := range images {
		name := image.Get("name")
		result[idx] = name.String()
	}

	return result, nil
}

func (c *ExtendedClient) RemoveImage(image string, tag string) error {
	address, err := c.extendedAPIAddress()
	if err != nil {
		return err
	}

	toDelete := address + "/api/v1/registry/images/" + image
	if tag != "" {
		toDelete += "?tag=" + url.QueryEscape(tag)
	}

	// TODO: error handling from body/header
	_, err = c.delete(toDelete)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}

	return nil
}

func (c *ExtendedClient) extendedAPIAddress() (string, error) {
	strURL := c.cfg.Address + "/info"
	response, err := c.get(strURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to get /info")
	}

	return "https://" + gjson.Get(response, "extended_api").String(), nil
}

func (c *ExtendedClient) get(urlStr string) (string, error) {
	c.logger.Trace().Str("url", urlStr).Msg("extended api get start")

	httpClient := http.Client{}
	httpClient.Transport = c.transport

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

	response, err := httpClient.Do(req)
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

func (c *ExtendedClient) delete(urlStr string) (string, error) {
	c.logger.Trace().Str("url", urlStr).Msg("extended api delete start")

	httpClient := http.Client{}
	httpClient.Transport = c.transport

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

	response, err := httpClient.Do(req)
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
