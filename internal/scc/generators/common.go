package generators

import (
	"os"

	"github.com/pkg/errors"
)

type Config struct {
	Server string
	Repo   string
	Token  string
	User   string
}

func FileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}
