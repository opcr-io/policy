package runtime_test

import (
	"testing"

	"github.com/opcr-io/policy/internal/runtime"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/stretchr/testify/assert"
)

func TestLoadManifest(t *testing.T) {
	assert := assert.New(t)

	tcs := []struct {
		input   string
		version ast.RegoVersion
	}{
		{"../../tests/fixtures/policy_v0/.manifest", ast.RegoV0},
		{"../../tests/fixtures/policy_v0v1/.manifest", ast.RegoV1},
		{"../../tests/fixtures/policy_v1/.manifest", ast.RegoV1},
	}

	for i, tc := range tcs {
		manifest, err := runtime.LoadManifest(tc.input)
		if err != nil {
			t.Logf("err: %v", err)
			t.Fail()
		}

		astRegoVersion := ast.RegoV0
		if manifest.RegoVersion != nil {
			astRegoVersion = ast.RegoVersionFromInt(*manifest.RegoVersion)
		}

		assert.Equal(tc.version, astRegoVersion, "tc%3d", i)
	}
}
