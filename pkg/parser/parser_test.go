package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculatePolicyRef(t *testing.T) {
	assert := require.New(t)

	defaultDomain := "opcr.io"
	expectedValues := map[string]string{
		"foo/bar:baz":           defaultDomain + "/foo/bar:baz",
		"foo/bar":               defaultDomain + "/foo/bar:latest",
		"docker.io/foo/bar:baz": "docker.io/foo/bar:baz",
	}

	for userRef, ref := range expectedValues {
		computedRef, err := CalculatePolicyRef(userRef, defaultDomain)
		if err != nil {
			assert.FailNow(err.Error())
		}
		assert.Equal(ref, computedRef)
	}

}
