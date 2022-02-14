package app

import (
	"io/ioutil"
	"testing"

	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/stretchr/testify/require"
)

var refExamples = []struct {
	userRef             string
	expectedInternalRef string
}{
	{"foo/bar:baz", "opcr.io/foo/bar:baz"},
	{"docker.io/foo/bar:baz", "docker.io/foo/bar:baz"},
	{"foo/bar", "opcr.io/foo/bar:latest"},
}

func TestRefCalculation(t *testing.T) {
	for _, tc := range refExamples {
		userRef := tc.userRef
		expectedInternalRef := tc.expectedInternalRef

		t.Run(tc.userRef, func(t *testing.T) {
			assert := require.New(t)

			p, cleanup, err := BuildTestPolicyApp(ioutil.Discard, ioutil.Discard, config.Path(""), func(c *config.Config) {})
			assert.NoError(err)
			defer cleanup()

			calculatedRef, err := p.calculatePolicyRef(userRef)
			assert.NoError(err)

			assert.Equal(expectedInternalRef, calculatedRef)
		})
	}

}
