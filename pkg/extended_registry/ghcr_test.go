package extendedregistry

import (
	"net/http"
	"os"
	"testing"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"github.com/golang/mock/gomock"
	mocksources "github.com/opcr-io/policy/pkg/extended_registry/mocks"
	"github.com/rs/zerolog"
	"gopkg.in/h2non/gock.v1"
	"gotest.tools/assert"
)

const (
	username = "testuser"
	password = "testpass"
)

var listResponse = `
[
	{
	  "id": 197,
	  "name": "hello_docker",
	  "package_type": "container",
	  "owner": {
		"login": "monalisa",
		"id": 9919,
		"node_id": "MDEyOk9yZ2FuaXphdGlvbjk5MTk=",
		"avatar_url": "https://avatars.monalisausercontent.com/u/9919?v=4",
		"gravatar_id": "",
		"url": "https://api.github.com/users/monalisa",
		"html_url": "https://github.com/github",
		"followers_url": "https://api.github.com/users/github/followers",
		"following_url": "https://api.github.com/users/github/following{/other_user}",
		"gists_url": "https://api.github.com/users/github/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/github/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/github/subscriptions",
		"organizations_url": "https://api.github.com/users/github/orgs",
		"repos_url": "https://api.github.com/users/github/repos",
		"events_url": "https://api.github.com/users/github/events{/privacy}",
		"received_events_url": "https://api.github.com/users/github/received_events",
		"type": "User",
		"site_admin": false
	  },
	  "version_count": 1,
	  "visibility": "private",
	  "url": "https://api.github.com/orgs/github/packages/container/hello_docker",
	  "created_at": "2020-05-19T22:19:11Z",
	  "updated_at": "2020-05-19T22:19:11Z",
	  "html_url": "https://github.com/orgs/github/packages/container/package/hello_docker"
	},
	{
	  "id": 198,
	  "name": "goodbye_docker",
	  "package_type": "container",
	  "owner": {
		"login": "github",
		"id": 9919,
		"node_id": "MDEyOk9yZ2FuaXphdGlvbjk5MTk=",
		"avatar_url": "https://avatars.githubusercontent.com/u/9919?v=4",
		"gravatar_id": "",
		"url": "https://api.github.com/users/monalisa",
		"html_url": "https://github.com/github",
		"followers_url": "https://api.github.com/users/github/followers",
		"following_url": "https://api.github.com/users/github/following{/other_user}",
		"gists_url": "https://api.github.com/users/github/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/github/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/github/subscriptions",
		"organizations_url": "https://api.github.com/users/github/orgs",
		"repos_url": "https://api.github.com/users/github/repos",
		"events_url": "https://api.github.com/users/github/events{/privacy}",
		"received_events_url": "https://api.github.com/users/github/received_events",
		"type": "User",
		"site_admin": false
	  },
	  "version_count": 2,
	  "visibility": "private",
	  "url": "https://api.github.com/user/monalisa/packages/container/goodbye_docker",
	  "created_at": "2020-05-20T22:19:11Z",
	  "updated_at": "2020-05-20T22:19:11Z",
	  "html_url": "https://github.com/user/monalisa/packages/container/package/goodbye_docker"
	}
  ]`

var listVersions = `
  [
	{
	  "id": 45763,
	  "name": "sha256:08a44bab0bddaddd8837a8b381aebc2e4b933768b981685a9e088360af0d3dd9",
	  "url": "https://api.github.com/users/octocat/packages/container/hello_docker/versions/45763",
	  "package_html_url": "https://github.com/users/octocat/packages/container/package/hello_docker",
	  "created_at": "2020-09-11T21:56:40Z",
	  "updated_at": "2021-02-05T21:32:32Z",
	  "html_url": "https://github.com/users/octocat/packages/container/hello_docker/45763",
	  "metadata": {
		"package_type": "container",
		"container": {
		  "tags": [
			"0.0.1"
		  ]
		}
	  }
	},
	{
	  "id": 881,
	  "name": "sha256:b3d3e366b55f9a54599220198b3db5da8f53592acbbb7dc7e4e9878762fc5344",
	  "url": "https://api.github.com/users/octocat/packages/container/hello_docker/versions/881",
	  "package_html_url": "https://github.com/users/octocat/packages/container/package/hello_docker",
	  "created_at": "2020-05-21T22:22:20Z",
	  "updated_at": "2021-02-05T21:32:32Z",
	  "html_url": "https://github.com/users/octocat/packages/container/hello_docker/881",
	  "metadata": {
		"package_type": "container",
		"container": {
		  "tags": [
			  "latest"
		  ]
		}
	  }
	}
  ]`

func TestGHCRListOrgs(t *testing.T) {
	testlog := zerolog.New(os.Stdout)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocksources.NewMockSource(ctrl)
	// mock return two orgs
	m.EXPECT().ListOrgs(gomock.Any(), gomock.Any(), &api.PaginationRequest{Size: 1, Token: ""}).
		Return([]string{"one"}, &api.PaginationResponse{NextToken: "test"}, nil)
	m.EXPECT().ListOrgs(gomock.Any(), gomock.Any(), &api.PaginationRequest{Size: 1, Token: "test"}).
		Return([]string{"two"}, &api.PaginationResponse{NextToken: ""}, nil)

	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)
	client.(*GHCRClient).sccClient = m
	var results *registry.ListOrgsResponse
	orgs, err := client.ListOrgs(&api.PaginationRequest{Size: 1, Token: ""})
	results = orgs
	for {
		if orgs.Page != nil {
			if orgs.Page.NextToken != "" {
				// Test pagination by taking 1 org at a time
				orgs, err = client.ListOrgs(&api.PaginationRequest{Size: 1, Token: orgs.Page.NextToken})
				results.Orgs = append(results.Orgs, orgs.Orgs...)
			} else {
				break
			}
		} else {
			break
		}
	}
	t.Log(results)
	assert.NilError(t, err)
	assert.Equal(t, len(results.Orgs), 2)
}

func TestGHCRListTags(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://api.github.com/user/packages/container/hello_docker/versions").Get("").Reply(200).BodyString(listVersions)

	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)

	resp, _, err := client.ListTags("", "hello_docker", &api.PaginationRequest{Size: -1, Token: ""}, false)
	assert.NilError(t, err)
	assert.Equal(t, len(resp), 2)
}
func TestGHCRGetTag(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://api.github.com/user/packages/container/hello_docker/versions").Get("").Reply(200).BodyString(listVersions)
	gock.New("https://api.github.com/users/octocat/packages/container/hello_docker/versions/45763").Get("").
		Reply(200).
		BodyString(`{
		"id": 45763,
		"name": "sha256:08a44bab0bddaddd8837a8b381aebc2e4b933768b981685a9e088360af0d3dd9",
		"url": "https://api.github.com/users/octocat/packages/container/hello_docker/versions/45763",
		"package_html_url": "https://github.com/users/octocat/packages/container/package/hello_docker",
		"created_at": "2020-09-11T21:56:40Z",
		"updated_at": "2021-02-05T21:32:32Z",
		"html_url": "https://github.com/users/octocat/packages/container/hello_docker/45763",
		"metadata": {
		  "package_type": "container",
		  "container": {
			"tags": [
			  "0.0.1"
			]
		  }
		}
	  }`)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)
	tag, err := client.GetTag("", "hello_docker", "0.0.1")
	assert.NilError(t, err)
	assert.Equal(t, tag.Name, "0.0.1")
}

func TestGHCRList(t *testing.T) {
	testlog := zerolog.New(os.Stdout)

	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://api.github.com/user/packages?package_type=container").
		Get("").Persist().
		Reply(200).BodyString(listResponse)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocksources.NewMockSource(ctrl)
	// mock return two orgs
	m.EXPECT().ListOrgs(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]string{"one"}, nil, nil)

	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)
	client.(*GHCRClient).sccClient = m
	orgs, err := client.ListOrgs(&api.PaginationRequest{Size: -1, Token: ""})
	assert.NilError(t, err)
	orgs.Orgs = append(orgs.Orgs, &api.RegistryOrg{Name: username}) // allow to get user packages
	response := &registry.ListImagesResponse{}
	for i := range orgs.Orgs {
		orgimages, _, err := client.ListRepos(orgs.Orgs[i].Name, &api.PaginationRequest{Size: -1, Token: ""})
		assert.NilError(t, err)
		response.Images = append(response.Images, orgimages.Images...)
	}
	assert.NilError(t, err)
	assert.Equal(t, len(response.Images) > 0, true)
}

func TestGHCRRemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)

	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://api.github.com/user/packages/container/test2/versions").
		Get("").
		Reply(200).BodyString(listVersions)
	gock.New("https://api.github.com/user/packages/container/test2/versions/4576").
		Delete("").
		Reply(204)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)

	err := client.RemoveImage("", "test2", "0.0.1")
	assert.NilError(t, err)
}
