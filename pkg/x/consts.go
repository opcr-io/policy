package x

import "os"

const IsWindows = `windows`

const (
	FileMode0600 os.FileMode = 0o600
	FileMode0700 os.FileMode = 0o700
)

const (
	VerbosityError int = iota
	VerbosityInfo
	VerbosityDebug
	VerbosityTrace
)
