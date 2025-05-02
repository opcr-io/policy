package app

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/aserto-dev/scc-lib/generators"
	"github.com/magefile/mage/sh"
	"github.com/opcr-io/policy/pkg/x"
	"github.com/opcr-io/policy/templates"

	"github.com/pkg/errors"
)

const (
	ciTemplateOrganization = "aserto-content"
	ctTemplateRepo         = "policy-template"
	ciTemplateTag          = "latest"
)

// Init: path is rootpath of project.
func (c *PolicyApp) Init(path, user, server, repo, scc, token string, overwrite, noSrc bool) error {
	defer c.Cancel()

	if !strings.EqualFold(scc, "github") && !strings.EqualFold(scc, "gitlab") {
		return errors.Errorf("not supported source code provider '%s'", scc)
	}

	if _, _, ok := strings.Cut(repo, "/"); !ok {
		return errors.New("invalid repo name, not org/repo")
	}

	if err := c.validatePath(path); err != nil {
		return err
	}

	generatorConfig := &generators.Config{
		Server: server,
		Repo:   repo,
		Token:  token,
		User:   user,
	}

	if err := c.generateContent(
		generatorConfig,
		path,
		scc,
		overwrite); err != nil {
		return err
	}

	if !noSrc {
		return c.generateContent(
			generatorConfig,
			path,
			fmt.Sprintf("%s/%s:%s", ciTemplateOrganization, ctTemplateRepo, ciTemplateTag),
			overwrite)
	}

	return nil
}

func (c *PolicyApp) generateContent(generatorConf *generators.Config, outPath, scc string, overwrite bool) error {
	prog := c.UI.Progressf("Generating files")
	prog.Start()

	templateRoot, err := fs.Sub(templates.Assets(), scc)
	if err != nil {
		return errors.Wrapf(err, "failed tog get sub fs %s", scc)
	}

	generator, err := generators.NewGenerator(
		generatorConf,
		c.Logger,
		templateRoot,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize generator")
	}

	if err := generator.Generate(outPath, overwrite); err != nil {
		return errors.Wrap(err, "failed to generate ci files")
	}

	prog.Stop()

	c.UI.Normal().Msg("The template was generated successfully.")

	return nil
}

func (c *PolicyApp) validatePath(path string) error {
	if exist, _ := generators.DirExist(path); !exist {
		if err := os.MkdirAll(path, x.FileMode0700); err != nil {
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
