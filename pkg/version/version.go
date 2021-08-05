package version

import (
	"fmt"
	"runtime"
	"time"
)

// values set by linker using ldflag -X
var (
	ver    string // nolint:gochecknoglobals // set by linker
	date   string // nolint:gochecknoglobals // set by linker
	commit string // nolint:gochecknoglobals // set by linker
)

// Info - version info.
type Info struct {
	Version string
	Date    string
	Commit  string
}

// GetInfo gets version stamp information.
func GetInfo() Info {
	if ver == "" {
		ver = "0.0.0"
	}

	if date == "" {
		date = time.Now().Format(time.RFC3339)
	}

	if commit == "" {
		commit = "undefined"
	}

	return Info{
		Version: ver,
		Date:    date,
		Commit:  commit,
	}
}

// String() returns the version info string.
func (vi Info) String() string {
	return fmt.Sprintf("%s g%s %s-%s [%s]",
		vi.Version,
		vi.Commit,
		runtime.GOOS,
		runtime.GOARCH,
		vi.Date,
	)
}
