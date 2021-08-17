package app

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

func (c *PolicyApp) Push(userRef string) error {
	defer c.Cancel()

	ref, err := c.calculatePolicyRef(userRef)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
	if err != nil {
		return err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	refs := ociStore.ListReferences()

	refDescriptor, ok := refs[ref]
	if !ok {
		return errors.Errorf("policy [%s] not found in the local store", ref)
	}

	c.UI.Normal().
		WithStringValue("digest", refDescriptor.Digest.String()).
		Msgf("Resolved ref [%s].", ref)

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: func(s string) ([]docker.RegistryHost, error) {
			caCert, err := ioutil.ReadFile("/home/toaster/.config/aserto/aserto-registry/certs/http-ca.crt")
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			client := &http.Client{}

			return []docker.RegistryHost{
				docker.RegistryHost{
					Host:         s,
					Scheme:       "https",
					Capabilities: docker.HostCapabilityPull | docker.HostCapabilityResolve | docker.HostCapabilityPush,
					Client:       client,
					Path:         "/v2",
					Authorizer: docker.NewDockerAuthorizer(
						docker.WithAuthClient(client),
						docker.WithAuthCreds(func(s string) (string, string, error) {
							creds, ok := c.Configuration.Servers[s]
							if !ok {
								return "", "", nil
							}

							return creds.Username, creds.Password, nil
						}),
						docker.WithAuthHeader(http.Header{
							"Authorization": []string{},
						})),
				},
			}, nil
		},
	})

	refDescriptor.Annotations[ocispec.AnnotationTitle] = ref

	pushDescriptor, err := oras.Push(c.Context, resolver, ref, ociStore, []ocispec.Descriptor{refDescriptor})
	if err != nil {
		return err
	}

	c.UI.Normal().
		WithStringValue("digest", pushDescriptor.Digest.String()).
		Msgf("Pushed ref [%s].", ref)

	return nil
}
