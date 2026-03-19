package tests_test

import (
	"path/filepath"
	"testing"

	"github.com/aserto-dev/runtime"

	"github.com/stretchr/testify/require"
)

type tc struct {
	PolicyName  string
	SourcePath  string
	SaveFile    string
	RegoVersion runtime.RegoVersion
}

var tcs = []tc{
	{
		PolicyName:  "ghcr.io/test/policy_v0:test",
		SourcePath:  "./fixtures/policy_v0",
		SaveFile:    "policy_v0.bundle.tar.gz",
		RegoVersion: runtime.RegoV0,
	},
	{
		PolicyName:  "ghcr.io/test/policy_v0v1:test",
		SourcePath:  "./fixtures/policy_v0v1",
		SaveFile:    "policy_v0v1.bundle.tar.gz",
		RegoVersion: runtime.RegoV0CompatV1,
	},
	{
		PolicyName:  "ghcr.io/test/policy_v1:test",
		SourcePath:  "./fixtures/policy_v1",
		SaveFile:    "policy_v1.bundle.tar.gz",
		RegoVersion: runtime.RegoV1,
	},
}

func TestBuild(t *testing.T) {
	require.DirExists(t, "./fixtures")

	for _, tc := range tcs {
		t.Run(tc.PolicyName, testBuild(&tc))
	}
}

func testBuild(tc *tc) func(*testing.T) {
	return func(t *testing.T) {
		cmdCtx := NewCmdContext(t)
		cleanup := cmdCtx.Setup()
		t.Cleanup(cleanup)

		sourcePath := []string{tc.SourcePath}

		policyName := tc.PolicyName

		fileName := filepath.Join("/tmp", tc.SaveFile)

		LogStep("version")
		require.NoError(t, NewVersionCmd(t).Run(cmdCtx))

		LogStep("images")
		require.NoError(t, NewImagesCmd(t).Run(cmdCtx))

		LogStep("build")
		require.NoError(t, NewBuildCmd(t,
			BuildWithTag(policyName),
			BuildWithSourcePath(sourcePath),
			BuildWithRegoVersion(tc.RegoVersion),
		).Run(cmdCtx))

		LogStep("inspect")
		require.NoError(t, NewInspectCmd(t,
			InspectWithPolicy(policyName),
		).Run(cmdCtx))

		LogStep("images")
		require.NoError(t, NewImagesCmd(t).Run(cmdCtx))

		LogStep("save")
		require.NoError(t, NewSaveCmd(t,
			SaveWithPolicy(policyName),
			SaveWithFile(fileName),
		).Run(cmdCtx))

		LogStep("rm")
		require.NoError(t, NewRmCmd(t,
			RmWithPolicies([]string{policyName}),
			RmWithForce(true),
		).Run(cmdCtx))

		LogStep("images")
		require.NoError(t, NewImagesCmd(t).Run(cmdCtx))
	}
}
