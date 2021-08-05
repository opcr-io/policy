package config

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

// DBStoreType represents the type of secret store to use
type DBStoreType int

const (
	// Bolt represents a secret store that saves data into a bolt db file
	Bolt DBStoreType = iota
)

func (t DBStoreType) String() string {
	return dbStoreToString[t]
}

var dbStoreToString = map[DBStoreType]string{
	Bolt: "bolt",
}

var dbStoreToID = map[string]DBStoreType{
	"bolt": Bolt,
}

// MarshalJSON marshals the enum as a quoted json string
func (t DBStoreType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(dbStoreToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON un-marshalls a quoted json string to the enum value
func (t *DBStoreType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	var ok bool
	*t, ok = dbStoreToID[j]
	if !ok {
		return errors.Errorf("'%s' is not a valid IDType", j)
	}
	return nil
}
