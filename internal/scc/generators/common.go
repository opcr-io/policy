package generators

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	gitDir    string = ".git"
	gitConfig string = "config"
)

type Config struct {
	Server string
	Repo   string
	Token  string
	User   string
}

func IsGitRepo(path string) error {
	if exist, err := DirExist(filepath.Join(path, gitDir)); !exist {
		return errors.Wrapf(err, "root path does not contain .git directory '%s'", path)
	}
	if exist, err := FileExist(filepath.Join(path, gitDir, gitConfig)); !exist {
		return errors.Wrapf(err, ".git directory does not contain config file '%s'", path)
	}
	return nil
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

func DirExist(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat directory '%s'", path)
	}
}
