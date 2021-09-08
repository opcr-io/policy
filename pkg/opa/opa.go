package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/compile"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/open-policy-agent/opa/util"
	"github.com/pkg/errors"
)

const stringType = "string"

type Builtin struct {
	Name string         `json:"name"`
	Decl types.Function `json:"decl"`
}

type Builtin1 Builtin
type Builtin2 Builtin
type Builtin3 Builtin
type Builtin4 Builtin
type BuiltinDyn Builtin

type BuiltinDefs struct {
	Builtin1   []Builtin1   `json:"builtin1,omitempty"`
	Builtin2   []Builtin2   `json:"builtin2,omitempty"`
	Builtin3   []Builtin3   `json:"builtin3,omitempty"`
	Builtin4   []Builtin4   `json:"builtin4,omitempty"`
	BuiltinDyn []BuiltinDyn `json:"builtinDyn,omitempty"`
}

func registerBuiltins(defs *BuiltinDefs) {
	for _, b := range defs.Builtin1 {
		rego.RegisterBuiltin1(&rego.Function{
			Name:    b.Name,
			Memoize: false,
			Decl:    &b.Decl,
		}, func(rego.BuiltinContext, *ast.Term) (*ast.Term, error) {
			return nil, nil
		})
	}

	for _, b := range defs.Builtin2 {
		rego.RegisterBuiltin2(&rego.Function{
			Name:    b.Name,
			Memoize: false,
			Decl:    &b.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
			return nil, nil
		})
	}

	for _, b := range defs.Builtin3 {
		rego.RegisterBuiltin3(&rego.Function{
			Name:    b.Name,
			Memoize: false,
			Decl:    &b.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2, op3 *ast.Term) (*ast.Term, error) {
			return nil, nil
		})
	}

	for _, b := range defs.Builtin4 {
		rego.RegisterBuiltin4(&rego.Function{
			Name:    b.Name,
			Memoize: false,
			Decl:    &b.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2, op3, op4 *ast.Term) (*ast.Term, error) {
			return nil, nil
		})
	}

	for _, b := range defs.BuiltinDyn {
		rego.RegisterBuiltinDyn(&rego.Function{
			Name:    b.Name,
			Memoize: false,
			Decl:    &b.Decl,
		}, func(bctx rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
			return nil, nil
		})
	}
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}

}

func generateAllBuiltins(paths []string) error {
	for _, path := range paths {
		manifestPath := filepath.Join(path, ".manifest")
		manifestExists, err := fileExists(manifestPath)
		if err != nil {
			return errors.Wrapf(err, "failed to determine if file [%s] exists", manifestPath)
		}

		if !manifestExists {
			continue
		}

		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			return errors.Wrapf(err, "failed to read manifest [%s]", manifestPath)
		}

		manifest := struct {
			Metadata struct {
				RequiredBuiltins *BuiltinDefs `json:"required_builtins"`
			} `json:"metadata,omitempty"`
		}{}
		err = json.Unmarshal(manifestBytes, &manifest)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal json from manifest [%s]", manifestPath)
		}

		if manifest.Metadata.RequiredBuiltins != nil {
			registerBuiltins(manifest.Metadata.RequiredBuiltins)
		}
	}

	return nil
}

func Build(buildParams *BuildParams, paths []string) error {
	err := generateAllBuiltins(paths)
	if err != nil {
		return err
	}

	return build(buildParams, paths)
}

// CapabilitiesFlag contains capabilities loaded based
// on build configuration flags
type CapabilitiesFlag struct {
	C    *ast.Capabilities
	Path string
}

// Type returns "string"
func (f *CapabilitiesFlag) Type() string {
	return stringType
}

// String returns the set path
func (f *CapabilitiesFlag) String() string {
	return f.Path
}

// Set loads capabilities from the provided path
func (f *CapabilitiesFlag) Set(s string) error {
	f.Path = s
	fd, err := os.Open(s)
	if err != nil {
		return err
	}
	defer fd.Close()
	f.C, err = ast.LoadCapabilitiesJSON(fd)
	return err
}

// RepeatedStringFlag is a flag that can be repeated
type RepeatedStringFlag struct {
	Values []string
	IsSet  bool
}

// Type returns "string"
func (f *RepeatedStringFlag) Type() string {
	return stringType
}

// String returns a comma-joined list of the flag values
func (f *RepeatedStringFlag) String() string {
	return strings.Join(f.Values, ",")
}

// Set appends a new flag value to the list
func (f *RepeatedStringFlag) Set(s string) error {
	f.Values = append(f.Values, s)
	f.IsSet = true
	return nil
}

// IsFlagSet returns true if at least one value has been added
func (f *RepeatedStringFlag) IsFlagSet() bool {
	return f.IsSet
}

// BuildParams contains all parameters used for doing a build
type BuildParams struct {
	Capabilities       *CapabilitiesFlag
	Target             *util.EnumFlag
	BundleMode         bool
	OptimizationLevel  int
	Entrypoints        RepeatedStringFlag
	OutputFile         string
	Revision           string
	Ignore             []string
	Debug              bool
	Algorithm          string
	Key                string
	Scope              string
	PubKey             string
	PubKeyID           string
	ClaimsFile         string
	ExcludeVerifyFiles []string
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

// Build builds a bundle using the OPA Runtime
func build(params *BuildParams, args []string) error {
	buf := bytes.NewBuffer(nil)

	// generate the bundle verification and signing config
	bvc, err := buildVerificationConfig(params.PubKey, params.PubKeyID, params.Algorithm, params.Scope, params.ExcludeVerifyFiles)
	if err != nil {
		return err
	}

	bsc := buildSigningConfig(params.Key, params.Algorithm, params.ClaimsFile)

	if bvc != nil || bsc != nil {
		if !params.BundleMode {
			return errors.Errorf("enable bundle mode (ie. --bundle) to verify or sign bundle files or directories")
		}
	}
	var capabilities *ast.Capabilities
	// if capabilities are not provided as a cmd flag,
	// then ast.CapabilitiesForThisVersion must be called
	// within dobuild to ensure custom builtins are properly captured
	if params.Capabilities.C != nil {
		capabilities = params.Capabilities.C
	} else {
		capabilities = ast.CapabilitiesForThisVersion()
	}

	compiler := compile.New().
		WithCapabilities(capabilities).
		WithTarget(params.Target.String()).
		WithAsBundle(params.BundleMode).
		WithOptimizationLevel(params.OptimizationLevel).
		WithOutput(buf).
		WithEntrypoints(params.Entrypoints.Values...).
		WithPaths(args...).
		WithFilter(buildCommandLoaderFilter(params.BundleMode, params.Ignore)).
		WithRevision(params.Revision).
		WithBundleVerificationConfig(bvc).
		WithBundleSigningConfig(bsc)

	if params.ClaimsFile == "" {
		compiler = compiler.WithBundleVerificationKeyID(params.PubKeyID)
	}

	err = compiler.Build(context.Background())
	if err != nil {
		return err
	}

	out, err := os.Create(params.OutputFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, buf)
	if err != nil {
		return err
	}

	return out.Close()
}

func buildCommandLoaderFilter(bundleMode bool, ignore []string) func(string, os.FileInfo, int) bool {
	return func(abspath string, info os.FileInfo, depth int) bool {
		if !bundleMode {
			if !info.IsDir() && strings.HasSuffix(abspath, ".tar.gz") {
				return true
			}
		}
		return loaderFilter{Ignore: ignore}.Apply(abspath, info, depth)
	}
}

func buildVerificationConfig(pubKey, pubKeyID, alg, scope string, excludeFiles []string) (*bundle.VerificationConfig, error) {
	if pubKey == "" {
		return nil, nil
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
