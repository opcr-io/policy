package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/scc-lib/generators"
	"github.com/magefile/mage/sh"
	"github.com/opcr-io/policy/pkg/policytemplates"

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

	names := strings.Split(repo, "/")
	if len(names) < 2 {
		return errors.New("invalid repo name, not org/repo")
	}

	err := c.validatePath(path)
	if err != nil {
		return err
	}

	generatorConfig := &generators.Config{
		Server: server,
		Repo:   repo,
		Token:  token,
		User:   user,
	}

	err = c.generateContent(
		generatorConfig,
		path,
		fmt.Sprintf("%s/%s:%s", c.Configuration.CITemplates.Organization, scc, c.Configuration.CITemplates.Tag),
		overwrite)

	if err != nil {
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

func (c *PolicyApp) generateContent(generatorConf *generators.Config, outPath, imageRef string, overwrite bool) error {
	policyTemplatesCfg := policytemplates.Config{
		Server:     c.Configuration.CITemplates.Server,
		PolicyRoot: c.Configuration.PoliciesRoot(),
	}

	prog := c.UI.Progressf("Generating files")
	prog.Start()

	ciTemplates := policytemplates.NewOCI(c.Context, c.Logger, c.TransportWithTrustedCAs(), policyTemplatesCfg)

	templateFs, err := ciTemplates.Load(imageRef)
	if err != nil {
		return errors.Wrapf(err, "failed to load '%s' template", imageRef)
	}

	generator, err := generators.NewGenerator(
		generatorConf,
		c.Logger,
		templateFs,
	)

	if err != nil {
		return errors.Wrap(err, "failed to initialize generator")
	}

	err = generator.Generate(outPath, overwrite)
	if err != nil {
		return errors.Wrap(err, "failed to generate ci files")
	}
	prog.Stop()

	c.UI.Normal().Msg("The template was generated successfully.")

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
