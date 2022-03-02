package app

import (
	"os"
	"strings"

	"github.com/magefile/mage/sh"

	"github.com/opcr-io/policy/pkg/generators"
	"github.com/pkg/errors"
)

// Init
// path is rootpath of project
func (c *PolicyApp) Init(path, user, server, repo, scc, token string, overwrite, noSrc bool) error {
	defer c.Cancel()

	if !strings.EqualFold(scc, "github") && !strings.EqualFold(scc, "gitlab") {
		return errors.Errorf("not supported source code provider '%s'", scc)
	}

	names := strings.Split(repo, "/")
	if len(names) < 2 {
		return errors.New("invalid repo name, not org/repo")
	}

	err := c.validatePath(path)
	if err != nil {
		return err
	}

	var sccGenerator generators.Generator
	sccStruct := &generators.SCC{
		Path:   path,
		Server: server,
		User:   user,
		Repo:   repo,
		Token:  token,
		UI:     c.UI,
	}
	switch scc {
	case "github":
		sccGenerator = generators.NewGithub(sccStruct)
	case "gitlab":
		sccGenerator = generators.NewGitlab(sccStruct)
	}

	if err := sccGenerator.Generate(overwrite); err != nil {
		return err
	}

	if noSrc {
		return nil
	}

	opa := generators.NewOpa(path, c.UI)
	if err := opa.Generate(overwrite); err != nil {
		return err
	}

	general := generators.NewGeneral(path, names[1], c.UI)
	if err := general.Generate(overwrite); err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) validatePath(path string) error {
	if exist, _ := generators.DirExist(path); !exist {
		if err := os.MkdirAll(path, 0700); err != nil {
			return errors.Errorf("root path not a directory '%s'", path)
		}
	}

	if err := generators.IsGitRepo(path); err != nil {
		if err := sh.RunV("git", "init", "--quiet", path); err != nil {
			return err
		}
		if err := generators.IsGitRepo(path); err != nil {
			return err
		}
	}
	return nil
}
