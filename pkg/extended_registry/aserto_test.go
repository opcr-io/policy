package extendedregistry

import (
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"
)

func Test_Aserto_List(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, &http.Transport{})
	images, err := client.ListRepos()
	assert.NilError(t, err)
	t.Log(images)
}
func Test_Aserto_SetVisibility(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, &http.Transport{})
	err := client.SetVisibility("dani/testpol", true)
	assert.NilError(t, err)
}

func Test_Aserto_RemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewAsertoClient(&testlog, &Config{Address: "https://opcr.io", Username: username, Password: password}, &http.Transport{})
	err := client.RemoveImage("dani/testpol", "latest")
	assert.NilError(t, err)
}
