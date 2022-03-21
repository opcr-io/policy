package extendedregistry

import (
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var opcrInfo = `
{"info":{"Version":"v0.1.2","Date":"2022-02-11T11:55:20Z","Commit":"8005fd3"},"extended_api":"api.opcr.io","grpc_extended_api":"api.opcr.io:8443"}
`

var listOrgs = `
{
	"orgs":[
{
	"Name":"test1"
},
{
	"Name":"test2"
}
	]
}
`

var listRepos = `
{
	"images":[
		{
			"Name":"test1",
			"Public":false
		},
		{
			"Name":"test2",
			"Public":true
		}
	]
}
`

func TestAsertoListOrgs(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://opcr.io").
		Get("/api/v1/registry/organizations").
		Reply(200).BodyString(listOrgs)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	orgs, _, err := client.ListOrgs(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(orgs.Orgs), 2)
}

func TestAsertoList(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://opcr.io").
		Get("/api/v1/registry/images").
		Reply(200).BodyString(listRepos)

	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	images, _, err := client.ListRepos("someorg", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(images.Images), 2)
}

func TestAsertoSetVisibility(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://api.opcr.io/api/v1/registry/images/dani/testpol/visibility").
		Post("").Reply(200)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	err := client.SetVisibility("dani", "testpol", true)
	assert.NoError(t, err)
}

func TestAsertoRemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://api.opcr.io/api/v1/registry/images/dani/testpol?tag=latest").
		Delete("").Reply(200)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	err := client.RemoveImage("dani", "testpol", "latest")
	assert.NoError(t, err)
}
