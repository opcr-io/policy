package runtime

import (
	"github.com/open-policy-agent/opa/v1/bundle"
)

type Manifest struct {
	bundle.Manifest

	RequiredBuiltIns StubBuiltinDefs
}

type MetadataEx struct {
	Metadata struct {
		RequiredBuiltins *StubBuiltinDefs `json:"required_builtins"`
	} `json:"metadata"`
}
