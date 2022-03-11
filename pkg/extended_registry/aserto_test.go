package extendedregistry

import (
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"gopkg.in/h2non/gock.v1"
	"gotest.tools/assert"
)

var opcrInfo = `
{"info":{"Version":"v0.1.2","Date":"2022-02-11T11:55:20Z","Commit":"8005fd3"},"extended_api":"api.opcr.io","grpc_extended_api":"api.opcr.io:8443"}
`

func Test_Aserto_ListOrgs(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	orgs, err := client.ListOrgs()
	assert.NilError(t, err)
	t.Log(orgs)
}
func Test_Aserto_List(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://opcr.io").
		Get("/api/v1/registry/images").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	images, err := client.ListRepos("")
	assert.NilError(t, err)
	t.Log(images)
}
func Test_Aserto_SetVisibility(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://api.opcr.io/api/v1/registry/images/dani/testpol/visibility").
		Post("").Reply(200)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	err := client.SetVisibility("dani/testpol", true)
	assert.NilError(t, err)
}

func Test_Aserto_RemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://opcr.io").
		Get("/info").
		Reply(200).BodyString(opcrInfo)
	gock.New("https://api.opcr.io/api/v1/registry/images/dani/testpol?tag=latest").
		Delete("").Reply(200)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, http.DefaultClient)
	err := client.RemoveImage("dani/testpol", "latest")
	assert.NilError(t, err)
}
