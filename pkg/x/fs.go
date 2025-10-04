package x

import "os"

const (
	OwnerReadWrite        os.FileMode = 0o600
	OwnerReadWriteExecute os.FileMode = 0o700
)
