package app_test

import (
	"testing"

	"github.com/opcr-io/policy/parser"
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

			calculatedRef, err := parser.CalculatePolicyRef(userRef, "opcr.io")
			assert.NoError(err)

			assert.Equal(expectedInternalRef, calculatedRef)
		})
	}

}
