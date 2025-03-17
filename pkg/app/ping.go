package app

import (
	"net/http"
	"net/url"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Ping(server, username, password string) error {
	defer c.Cancel()

	client := &http.Client{Transport: c.TransportWithTrustedCAs()}

	authorizer := docker.NewDockerAuthorizer(
		docker.WithAuthClient(client),
		docker.WithAuthCreds(func(s string) (string, string, error) {
			return username, password, nil
		}),
	)

	// Request #1 for login.
	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "https",
			Host:   server,
			Path:   "/v2/",
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to ping server [%s]", server)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.UI.Problem().WithErr(err).Msg("failed to close response body")
		}
	}()

	if err = authorizer.AddResponses(c.Context, []*http.Response{resp}); err != nil {
		return errors.Wrapf(err, "failed to consume response from server [%s]", server)
	}

	// Request #2 (with authentication).
	req2 := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "https",
			Host:   server,
			Path:   "/v2/",
		},
		Header: http.Header{},
	}
	err = authorizer.Authorize(c.Context, req2)
	if err != nil {
		return errors.Wrapf(err, "failed to authorize request for server [%s]", server)
	}
	resp2, err := client.Do(req2)
	if err != nil {
		return errors.Wrapf(err, "failed to login to server [%s]", server)
	}
	defer func() {
		if err := resp2.Body.Close(); err != nil {
			c.UI.Problem().WithErr(err).Msg("failed to close response body")
		}
	}()

	if resp2.StatusCode != http.StatusOK {
		return errors.Errorf("authentication to server [%s] failed, status [%s]", server, resp.Status)
	}

	return nil
}
