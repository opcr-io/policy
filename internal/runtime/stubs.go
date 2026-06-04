package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/pkg/errors"
)

type StubBuiltin struct {
	Name string         `json:"name"`
	Decl types.Function `json:"decl"`
}

type (
	StubBuiltin1   StubBuiltin
	StubBuiltin2   StubBuiltin
	StubBuiltin3   StubBuiltin
	StubBuiltin4   StubBuiltin
	StubBuiltinDyn StubBuiltin
)

type StubBuiltinDefs struct {
	Builtin1   []StubBuiltin1   `json:"builtin1,omitempty"`
	Builtin2   []StubBuiltin2   `json:"builtin2,omitempty"`
	Builtin3   []StubBuiltin3   `json:"builtin3,omitempty"`
	Builtin4   []StubBuiltin4   `json:"builtin4,omitempty"`
	BuiltinDyn []StubBuiltinDyn `json:"builtin_dyn,omitempty"`
}

func generateAllStubBuiltins(paths []string) error {
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

		manifest := MetadataEx{}

		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return errors.Wrapf(err, "failed to unmarshal json from manifest [%s]", manifestPath)
		}

		if manifest.Metadata.RequiredBuiltins != nil {
			registerStubBuiltins(manifest.Metadata.RequiredBuiltins)
		}
	}

	return nil
}

func RegisterStubBuiltins(defs *StubBuiltinDefs) {
	registerStubBuiltins(defs)
}

func registerStubBuiltins(defs *StubBuiltinDefs) {
	for _, b := range defs.Builtin1 {
		builtin := b

		if topdown.GetBuiltin(b.Name) != nil {
			continue
		}

		rego.RegisterBuiltin1(&rego.Function{
			Name:    builtin.Name,
			Memoize: false,
			Decl:    &builtin.Decl,
		}, func(rego.BuiltinContext, *ast.Term) (*ast.Term, error) {
			return ast.NullTerm(), nil
		})
	}

	for _, b := range defs.Builtin2 {
		builtin := b

		if topdown.GetBuiltin(b.Name) != nil {
			continue
		}

		rego.RegisterBuiltin2(&rego.Function{
			Name:    builtin.Name,
			Memoize: false,
			Decl:    &builtin.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
			return ast.NullTerm(), nil
		})
	}

	for _, b := range defs.Builtin3 {
		builtin := b

		if topdown.GetBuiltin(b.Name) != nil {
			continue
		}

		rego.RegisterBuiltin3(&rego.Function{
			Name:    builtin.Name,
			Memoize: false,
			Decl:    &builtin.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2, op3 *ast.Term) (*ast.Term, error) {
			return ast.NullTerm(), nil
		})
	}

	for _, b := range defs.Builtin4 {
		builtin := b

		if topdown.GetBuiltin(b.Name) != nil {
			continue
		}

		rego.RegisterBuiltin4(&rego.Function{
			Name:    builtin.Name,
			Memoize: false,
			Decl:    &builtin.Decl,
		}, func(bctx rego.BuiltinContext, op1, op2, op3, op4 *ast.Term) (*ast.Term, error) {
			return ast.NullTerm(), nil
		})
	}

	for _, b := range defs.BuiltinDyn {
		builtin := b

		if topdown.GetBuiltin(b.Name) != nil {
			continue
		}

		rego.RegisterBuiltinDyn(&rego.Function{
			Name:    builtin.Name,
			Memoize: false,
			Decl:    &builtin.Decl,
		}, func(bctx rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
			return ast.NullTerm(), nil
		})
	}
}
