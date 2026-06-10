package runtime

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/compile"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/pkg/errors"
)

// BuildTargetType represents the type of build target.
type BuildTargetType int

const (
	Rego BuildTargetType = iota
	Wasm
)

func (t BuildTargetType) String() string {
	return buildTargetTypeToString[t]
}

var buildTargetTypeToString = map[BuildTargetType]string{
	Rego: "rego",
	Wasm: "wasm",
}

type RegoVersion int

const DefaultRegoVersion = RegoV1

const (
	RegoUndefined RegoVersion = iota
	// RegoV0 is the default, original Rego syntax.
	RegoV0
	// RegoV0CompatV1 requires modules to comply with both the RegoV0 and RegoV1 syntax (as when 'rego.v1' is imported in a module).
	// Shortly, RegoV1 compatibility is required, but 'rego.v1' or 'future.keywords' must also be imported.
	RegoV0CompatV1
	// RegoV1 is the Rego syntax enforced by OPA 1.0; e.g.:
	// future.keywords part of default keyword set, and don't require imports;
	// 'if' and 'contains' required in rule heads;
	// (some) strict checks on by default.
	RegoV1
)

const (
	regoUndefined string = "undefined"
	regoV0        string = "rego.v0"
	regoV0V1      string = "rego.v0v1"
	regoV1        string = "rego.v1"
)

func (v RegoVersion) ToAstRegoVersion() ast.RegoVersion {
	switch v {
	case RegoUndefined:
		return ast.RegoUndefined
	case RegoV0:
		return ast.RegoV0
	case RegoV0CompatV1:
		return ast.RegoV0CompatV1
	case RegoV1:
		return ast.RegoV1
	default:
		return ast.RegoUndefined
	}
}

func (v RegoVersion) String() string {
	switch v {
	case RegoUndefined:
		return regoUndefined
	case RegoV0:
		return regoV0
	case RegoV0CompatV1:
		return regoV0V1
	case RegoV1:
		return regoV1
	default:
		return regoUndefined
	}
}

func RegoVersionFromString(v string) RegoVersion {
	switch v {
	case regoV0:
		return RegoV0
	case regoV0V1:
		return RegoV0CompatV1
	case regoV1:
		return RegoV1
	default:
		return RegoV1
	}
}

// BuildParams contains all parameters used for doing a build.
type BuildParams struct {
	CapabilitiesJSONFile string
	Target               BuildTargetType
	OptimizationLevel    int
	Entrypoints          []string
	OutputFile           string
	Revision             string
	Ignore               []string
	Debug                bool
	Algorithm            string
	Key                  string
	Scope                string
	PubKey               string
	PubKeyID             string
	ClaimsFile           string
	ExcludeVerifyFiles   []string
	RegoVersion          RegoVersion
}

// Build builds a policy bundle using OPA's compiler.
func (r *Runtime) Build(params *BuildParams, paths []string) error {
	buf := bytes.NewBuffer(nil)

	if err := generateAllStubBuiltins(paths); err != nil {
		return err
	}

	// generate the bundle verification and signing config.
	var (
		bvc *bundle.VerificationConfig
		err error
	)

	if params.PubKey != "" {
		bvc, err = buildVerificationConfig(params.PubKey, params.PubKeyID, params.Algorithm, params.Scope, params.ExcludeVerifyFiles)
		if err != nil {
			return err
		}
	}

	bsc := buildSigningConfig(params.Key, params.Algorithm, params.ClaimsFile)

	var capabilities *ast.Capabilities
	// if capabilities are not provided then ast.CapabilitiesForThisVersion must be used.
	if params.CapabilitiesJSONFile == "" {
		capabilities = ast.CapabilitiesForThisVersion()
	} else {
		capabilitiesJSON, err := os.ReadFile(params.CapabilitiesJSONFile)
		if err != nil {
			return errors.Wrapf(err, "couldn't read capabilities JSON file [%s]", params.CapabilitiesJSONFile)
		}

		capabilities, err = ast.LoadCapabilitiesJSON(bytes.NewBuffer(capabilitiesJSON))
		if err != nil {
			return errors.Wrapf(err, "failed to load capabilities file [%s]", params.CapabilitiesJSONFile)
		}
	}

	compiler := compile.New().
		WithCapabilities(capabilities).
		WithTarget(params.Target.String()).
		WithAsBundle(true).
		WithOptimizationLevel(params.OptimizationLevel).
		WithOutput(buf).
		WithEntrypoints(params.Entrypoints...).
		WithPaths(paths...).
		WithFilter(buildCommandLoaderFilter(true, params.Ignore)).
		WithRevision(params.Revision).
		WithBundleVerificationConfig(bvc).
		WithBundleSigningConfig(bsc).
		WithRegoVersion(params.RegoVersion.ToAstRegoVersion())

	if params.ClaimsFile == "" {
		compiler = compiler.WithBundleVerificationKeyID(params.PubKeyID)
	}

	if err := compiler.Build(context.Background()); err != nil {
		return err
	}

	out, err := os.Create(params.OutputFile)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, buf); err != nil {
		return err
	}

	return out.Close()
}

func buildCommandLoaderFilter(bundleMode bool, ignore []string) func(string, os.FileInfo, int) bool {
	return func(absPath string, info os.FileInfo, depth int) bool {
		if !bundleMode {
			if !info.IsDir() && strings.HasSuffix(absPath, ".tar.gz") {
				return true
			}
		}

		return loaderFilter{Ignore: ignore}.Apply(absPath, info, depth)
	}
}

func buildVerificationConfig(pubKey, pubKeyID, alg, scope string, excludeFiles []string) (*bundle.VerificationConfig, error) {
	if pubKey == "" {
		return nil, errors.New("pubKey is empty")
	}

	keyConfig := &bundle.KeyConfig{
		Key:       pubKey,
		Algorithm: alg,
		Scope:     scope,
	}

	return bundle.NewVerificationConfig(map[string]*bundle.KeyConfig{pubKeyID: keyConfig}, pubKeyID, scope, excludeFiles), nil
}

func buildSigningConfig(key, alg, claimsFile string) *bundle.SigningConfig {
	if key == "" {
		return nil
	}

	return bundle.NewSigningConfig(key, alg, claimsFile)
}

func fileExists(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

type loaderFilter struct {
	Ignore []string
}

func (f loaderFilter) Apply(abspath string, info os.FileInfo, depth int) bool {
	for _, s := range f.Ignore {
		if loader.GlobExcludeName(s, 1)(abspath, info, depth) {
			return true
		}
	}

	return false
}
