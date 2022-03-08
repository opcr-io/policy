package extendedregistry

import (
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"
)

const (
	username = "testusername"
	password = "testpassword"
)

func Test_GHCR_List(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, &http.Transport{})
	images, err := client.ListRepos()
	assert.NilError(t, err)
	assert.Equal(t, len(images), 3)
	testlog.Debug().Msgf("Received images: %v", images)
}

func Test_GHCR_RemoveImage(t *testing.T) {
	testlog := zerolog.New(os.Stdout)
	client := NewGHCRClient(&testlog, &Config{Address: "https://ghcr.io", Username: username, Password: password}, &http.Transport{})

	err := client.RemoveImage("test2", "0.0.1")
	assert.NilError(t, err)
}
