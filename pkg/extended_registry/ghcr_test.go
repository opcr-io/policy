package extendedregistry

import (
	"net/http"
	"os"
	"testing"

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
		  "tags": []
		}
	  }
	}
  ]`

func Test_GHCR_ListOrgs(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)

	orgs, err := client.ListOrgs()
	assert.NilError(t, err)
	t.Log(orgs)
}

func Test_GHCR_List(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	/*
		defer gock.Off() // Flush pending mocks after test execution

		gock.New("https://api.github.com/user/packages?package_type=container").
			Get("").Persist().
			Reply(200).BodyString(listResponse)
	*/
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)
	orgs, err := client.ListOrgs()
	assert.NilError(t, err)
	var images []*PolicyImage
	for i := range orgs {
		orgimages, err := client.ListRepos(orgs[i])
		assert.NilError(t, err)
		images = append(images, orgimages...)
	}
	assert.NilError(t, err)
	assert.Equal(t, len(images) > 0, true)
	testlog.Debug().Msgf("Received images: %v", images)
}

func Test_GHCR_RemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)

	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://api.github.com/user/packages/container/test2/versions").
		Get("").
		Reply(200).BodyString(listVersions)
	gock.New("https://api.github.com/user/packages/container/test2/versions/4576").
		Delete("").
		Reply(204)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, http.DefaultClient)

	err := client.RemoveImage("test2", "0.0.1")
	assert.NilError(t, err)
}
