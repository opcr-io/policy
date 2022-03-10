package extendedregistry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GHCRClient struct {
	base *xClient
}

func NewGHCRClient(logger *zerolog.Logger, cfg *Config, transport *http.Transport) ExtendedClient {
	client := NewExtendedClient(logger, cfg, transport)

	return &GHCRClient{
		base: client,
	}
}

func (g *GHCRClient) ListRepos() ([]*PolicyImage, error) {
	g.base.logger.Debug().Msg("List images")
	resp, err := g.base.get("https://api.github.com/user/packages?package_type=container")
	if err != nil {
		return nil, errors.Wrap(err, "failed to list container type packages from ghcr")
	}
	g.base.logger.Trace().Msgf("Response from api.github.com %v", resp)

	var images []struct {
		Name       string `json:"name"`
		Visibility string `json:"visibility"`
		Owner      struct {
			Login string `json:"login"`
		}
	}

	err = json.Unmarshal([]byte(resp), &images)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal policy list response")
	}

	response := make([]*PolicyImage, len(images))
	for i := range images {
		policy := PolicyImage{}
		policy.Name = images[i].Owner.Login + "/" + images[i].Name
		if images[i].Visibility == "public" {
			policy.Public = true
		} else {
			policy.Public = false
		}
		response[i] = &policy
	}
	return response, nil
}
func (g *GHCRClient) SetVisibility(image string, public bool) error {
	return errors.New("not implemented")
}

func (g *GHCRClient) GetTags(image string) (string, error) {
	resp, err := g.base.get(fmt.Sprintf("https://api.github.com/user/packages/%s/%s/versions", "container", image))
	if err != nil {
		return "", errors.Wrap(err, "failed to list container versions")
	}
	return resp, nil
}

func (g *GHCRClient) RemoveImage(image, tag string) error {
	resp, err := g.GetTags(image)
	if err != nil {
		return errors.Wrap(err, "failed to get tags")
	}
	type Container struct {
		List []string `json:"tags"`
	}

	type Metadata struct {
		PackageType string    `json:"package_type"`
		Tags        Container `json:"container"`
	}
	var versions []struct {
		ID   int      `json:"id"`
		Meta Metadata `json:"metadata"`
	}

	err = json.Unmarshal([]byte(resp), &versions)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal policy list versions")
	}
	var deleteid int
	for i := range versions {
		for j := range versions[i].Meta.Tags.List {
			if versions[i].Meta.Tags.List[j] == tag {
				deleteid = versions[i].ID
				break
			}
		}
	}

	deleteurl := fmt.Sprintf("https://api.github.com/user/packages/%s/%s/versions/%d", "container", image, deleteid)
	_, err = g.base.delete(deleteurl)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}
	return nil
}
