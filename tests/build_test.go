package tests_test

import (
	"path/filepath"
	"testing"

	"github.com/aserto-dev/runtime"

	"github.com/stretchr/testify/require"
)

func TestBuildV0(t *testing.T) {
	cmdCtx := NewCmdContext(t)
	cleanup := cmdCtx.Setup()
	t.Cleanup(cleanup)

	sourcePath := []string{"/Users/gertd/workspace/src/github.com/opcr-io/policy/tests/fixtures/policy_v0"}
	policyName := "test/policy_v0:test"
	fileName := filepath.Join(t.TempDir(), "policy_v0.bundle.tar.gz")

	require.NoError(t, NewVersionCmd(t).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewImagesCmd(t).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewBuildCmd(t,
		BuildWithTag(policyName),
		BuildWithSourcePath(sourcePath),
		BuildWithRegoVersion(runtime.RegoV0),
	).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewInspectCmd(t,
		InspectWithPolicy(policyName),
	).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewImagesCmd(t).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewSaveCmd(t,
		SaveWithPolicy(policyName),
		SaveWithFile(fileName),
	).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewRmCmd(t,
		RmWithPolicies([]string{policyName}),
		RmWithForce(true),
	).Run(cmdCtx))
	t.Log("\n")

	require.NoError(t, NewImagesCmd(t).Run(cmdCtx))
	t.Log("\n")
}

func TestBuildV0V1(t *testing.T) {
	cmdCtx := NewCmdContext(t)
	cleanup := cmdCtx.Setup()
	t.Cleanup(cleanup)

	err := NewBuildCmd(t,
		BuildWithTag("test/policy_v0v1:test"),
		BuildWithSourcePath([]string{"/Users/gertd/workspace/src/github.com/opcr-io/policy/tests/fixtures/policy_v0v1"}),
		BuildWithRegoVersion(runtime.RegoV0CompatV1),
	).Run(cmdCtx)

	require.NoError(t, err)
}

func TestBuildV1(t *testing.T) {
	cmdCtx := NewCmdContext(t)
	cleanup := cmdCtx.Setup()
	t.Cleanup(cleanup)

	err := NewBuildCmd(t,
		BuildWithTag("test/policy_v1:test"),
		BuildWithSourcePath([]string{"/Users/gertd/workspace/src/github.com/opcr-io/policy/tests/fixtures/policy_v1"}),
		BuildWithRegoVersion(runtime.RegoV1),
	).Run(cmdCtx)

	require.NoError(t, err)
}
