package app

import (
	"net/http"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Pull(userRef string) error {
	defer c.Cancel()

	ref, err := parser.CalculatePolicyRef(userRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	c.UI.Normal().
		WithStringValue("ref", userRef).
		Msg("Pulling.")

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	digest, err := ociClient.Pull(ref)
	if err != nil {
		return errors.Wrap(err, "oras pull failed")
	}

	c.UI.Normal().
		WithStringValue("digest", digest.String()).
		Msgf("Pulled ref [%s].", ref)

	return nil
}

func (c *PolicyApp) getHosts(server string) ([]docker.RegistryHost, error) {
	client := &http.Client{Transport: c.TransportWithTrustedCAs()}
	return []docker.RegistryHost{
		{
			Host:         server,
			Scheme:       "https",
			Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
			Client:       client,
			Path:         "/v2",
			Authorizer: docker.NewDockerAuthorizer(
				docker.WithAuthClient(client),
				docker.WithAuthCreds(func(s string) (string, string, error) {
					creds, ok := c.Configuration.Servers[s]
					if !ok || (creds.Username == "" && creds.Password == "") {
						return " ", " ", nil
					}

					return creds.Username, creds.Password, nil
				})),
		},
	}, nil
}
