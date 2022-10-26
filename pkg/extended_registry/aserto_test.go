package extendedregistry

import (
	"context"
	"testing"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/golang/mock/gomock"
	mocksources "github.com/opcr-io/policy/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAsertoListOrgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	regTestClient := mocksources.NewMockRegistryClient(ctrl)
	regTestClient.EXPECT().ListOrgs(gomock.Any(), gomock.Any()).Return(
		&registry.ListOrgsResponse{
			Orgs: []*api.RegistryOrg{
				{
					Name: "test",
				},
				{
					Name: "test2",
				},
			},
		}, nil,
	)
	client := &AsertoClient{registryClient: regTestClient}

	orgs, err := client.ListOrgs(context.Background(), &api.PaginationRequest{Size: -1, Token: ""})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orgs.Orgs))
}

func TestAsertoList(t *testing.T) {
	// Example aserto client to make a call to opcr.io
	// testlog := zerolog.New(os.Stdout)
	// client, err := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password, GRPCAddress: "api.opcr.io:8443"})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	regTestClient := mocksources.NewMockRegistryClient(ctrl)
	regTestClient.EXPECT().ListImages(gomock.Any(), gomock.Any()).Return(
		&registry.ListImagesResponse{
			Images: []*api.PolicyImage{
				{
					Name:   "some/test",
					Public: true,
				},
				{
					Name:   "some/test2",
					Public: false,
				},
			},
		}, nil,
	)
	client := &AsertoClient{registryClient: regTestClient}
	images, _, err := client.ListRepos(context.Background(), "some", nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(images.Images))
}

func TestAsertoSetVisibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	regTestClient := mocksources.NewMockRegistryClient(ctrl)
	regTestClient.EXPECT().SetImageVisibility(gomock.Any(), &registry.SetImageVisibilityRequest{
		Image:        "image",
		Organization: "org",
		Public:       true,
	}).Return(nil, nil)
	client := &AsertoClient{registryClient: regTestClient}

	err := client.SetVisibility(context.Background(), "org", "image", true)
	assert.NoError(t, err)
}

func TestAsertoRemoveImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	regTestClient := mocksources.NewMockRegistryClient(ctrl)
	regTestClient.EXPECT().RemoveImage(gomock.Any(), &registry.RemoveImageRequest{
		Image:        "testpol",
		Tag:          "latest",
		Organization: "org",
	}).Return(nil, nil)
	client := &AsertoClient{registryClient: regTestClient}

	err := client.RemoveImage(context.Background(), "org", "testpol", "latest")
	assert.NoError(t, err)
}
