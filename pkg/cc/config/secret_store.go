package config

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

// SecretStoreType represents the type of secret store to use
type SecretStoreType int

const (
	// InMemory represents a secret store that saves data in memory
	InMemory SecretStoreType = iota
	// Vault represents a secret store that uses Hashicorp Vault
	Vault
)

func (t SecretStoreType) String() string {
	return secretStoreToString[t]
}

var secretStoreToString = map[SecretStoreType]string{
	InMemory: "memory",
	Vault:    "vault",
}

var secretStoreToID = map[string]SecretStoreType{
	"memory": InMemory,
	"vault":  Vault,
}

// MarshalJSON marshals the enum as a quoted json string
func (t SecretStoreType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(secretStoreToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON un-marshalls a quoted json string to the enum value
func (t *SecretStoreType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	var ok bool
	*t, ok = secretStoreToID[j]
	if !ok {
		return errors.Errorf("'%s' is not a valid IDType", j)
	}
	return nil
}
