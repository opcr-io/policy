package extendedregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GHCRClient struct {
	base *xClient
}

var images []struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	Owner      struct {
		Login string `json:"login"`
	}
}

func NewGHCRClient(logger *zerolog.Logger, cfg *Config, client *http.Client) ExtendedClient {
	baseClient := newExtendedClient(logger, cfg, client)

	return &GHCRClient{
		base: baseClient,
	}
}

func (g *GHCRClient) ListOrgs() ([]string, error) {
	orgsresponse, err := g.base.get("https://api.github.com/user/orgs")
	if err != nil {
		return nil, errors.Wrap(err, "could not list organizations")
	}
	var orgs []struct {
		Login string `json:"login"`
	}
	err = json.Unmarshal([]byte(orgsresponse), &orgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}
	var response []string
	for i := range orgs {
		response = append(response, orgs[i].Login)
	}
	return response, nil
}

func (g *GHCRClient) ListRepos(org string) ([]*PolicyImage, error) {
	g.base.logger.Debug().Msg("List images")
	var resp string
	var err error
	if org != "" {
		resp, err = g.base.get(fmt.Sprintf("https://api.github.com/orgs/%s/packages?package_type=container", org))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list containers for org %s", org)
		}
	} else {
		resp, err = g.base.get("https://api.github.com/user/packages?package_type=container")
		if err != nil {
			return nil, errors.Wrap(err, "failed to list container type packages from ghcr")
		}
	}

	err = json.Unmarshal([]byte(resp), &images)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal policy list response")
	}
	var response []*PolicyImage

	for i := range images {
		policy := PolicyImage{}
		policy.Name = images[i].Owner.Login + "/" + images[i].Name
		if images[i].Visibility == "public" {
			policy.Public = true
		} else {
			policy.Public = false
		}

		tags, err := g.GetTags(images[i].Name, org)
		if err != nil {
			g.base.logger.Err(errors.Wrapf(err, "failed to load tags for image %s", images[i].Name))
		}
		if g.skipImage(tags, g.base.cfg.Address, &policy, g.base.cfg.Username, g.base.cfg.Password) {
			continue
		}

		response = append(response, &policy)
	}

	return response, nil
}

func (g *GHCRClient) SetVisibility(image string, public bool) error {
	return errors.New("please set the visibility using the web UI")
}

//GetTags returns tags on the image - if org is empty it returns the tags of the user's image
func (g *GHCRClient) GetTags(image string, org string) ([]string, error) {
	var resp string
	var err error
	if org != "" {
		resp, err = g.base.get(fmt.Sprintf("https://api.github.com/orgs/%s/packages/%s/%s/versions", org, "container", image))
		if err != nil {
			return nil, errors.Wrap(err, "failed to list container versions")
		}
	} else {
		resp, err = g.base.get(fmt.Sprintf("https://api.github.com/user/packages/%s/%s/versions", "container", image))
		if err != nil {
			return nil, errors.Wrap(err, "failed to list container versions")
		}
	}
	type ContainerInfo struct {
		Tags []string `json:"tags"`
	}
	type Metadata struct {
		PackageType string `json:"package_type"`
		Container   ContainerInfo
	}
	var tagResponse []struct {
		ID   int      `json:"ID"`
		Meta Metadata `json:"metadata"`
	}
	err = json.Unmarshal([]byte(resp), &tagResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal tag response for image %s", image)
	}

	var response []string
	for i := range tagResponse {
		response = append(response, tagResponse[i].Meta.Container.Tags...)
	}
	return response, nil
}

func (g *GHCRClient) RemoveImage(image, tag string) error {
	deleteid := 0 //TODO: add getVersionid

	deleteurl := fmt.Sprintf("https://api.github.com/user/packages/%s/%s/versions/%d", "container", image, deleteid)
	_, err := g.base.delete(deleteurl)
	if err != nil {
		return errors.Wrap(err, "failed to remove image")
	}
	return nil
}

func (g *GHCRClient) validImage(repoName, username, password string) bool {
	repo, err := name.ParseReference(repoName)
	if err != nil {
		g.base.logger.Err(err)
		return false
	}
	descriptor, err := remote.Head(repo,
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		}))
	if err != nil {
		g.base.logger.Err(err)
		return false
	}
	return strings.Contains(string(descriptor.MediaType), "vnd.oci.image.manifest")
}
func (g *GHCRClient) skipImage(tags []string, server string, image *PolicyImage, username, password string) bool {
	skipImage := false
	for i := range tags {
		image := fmt.Sprintf("%s/%s:%s", strings.Replace(server, "https://", "", -1), image.Name, tags[i])
		if !g.validImage(image, username, password) {
			skipImage = true
			break
		}
	}
	return skipImage
}
