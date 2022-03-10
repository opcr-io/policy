package extendedregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type AsertoClient struct {
	base *xClient
}

func NewAsertoClient(logger *zerolog.Logger, cfg *Config, client *http.Client) ExtendedClient {
	baseClient := NewExtendedClient(logger, cfg, client)

	return &AsertoClient{
		base: baseClient,
	}
}

func (c *AsertoClient) ListRepos() ([]*PolicyImage, error) {
	address, err := c.extendedAPIAddress()
	if err != nil {
		return nil, err
	}

	jsonBody, err := c.base.get(address + "/api/v1/registry/images")
	if err != nil {
		return nil, errors.Wrap(err, "failed to list images")
	}

	response := struct {
		Images []*PolicyImage `json:"images"`
	}{}

	err = json.Unmarshal([]byte(jsonBody), &response)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal policy list response")
	}

	return response.Images, nil
}
func (c *AsertoClient) SetVisibility(image string, public bool) error {
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
func (c *AsertoClient) RemoveImage(image, tag string) error {
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

func (c *AsertoClient) extendedAPIAddress() (string, error) {
	strURL := c.base.cfg.Address + "/info"
	response, err := c.base.get(strURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to get /info")
	}

	return "https://" + gjson.Get(response, "extended_api").String(), nil
}
