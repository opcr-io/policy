package app

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/aserto-dev/scc-lib/generators"
	"github.com/opcr-io/policy/templates"
	"github.com/pkg/errors"
)

func (c *PolicyApp) TemplateApply(name, outPath string, overwrite bool) error {
	defer c.Cancel()

	if name == "" {
		return errors.New("template name is required")
	}

	prog := c.UI.Progressf("Processing template '%s'", name)
	prog.Start()

	templatesInfo, err := c.listTemplates()
	if err != nil {
		return err
	}

	var tmplInfo templateInfo

	for _, ti := range templatesInfo {
		if ti.name == name {
			tmplInfo = ti
			break
		}
	}

	if tmplInfo.name == "" {
		return errors.Errorf("template '%s' not found", name)
	}

	prog.Stop()

	generatorCfg, err := c.getGeneratorConfig(tmplInfo)
	if err != nil {
		return err
	}

	prog.ChangeMessage("Generating files")
	prog.Start()
	templateRoot, err := fs.Sub(templates.Assets(), name)
	if err != nil {
		return errors.Wrapf(err, "failed tog get sub fs %s", name)
	}

	generator, err := generators.NewGenerator(
		&generatorCfg,
		c.Logger,
		templateRoot,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize generator")
	}

	err = generator.Generate(outPath, overwrite)
	if err != nil {
		return errors.Wrap(err, "failed to generate policy")
	}

	prog.Stop()

	c.UI.Normal().Msgf("The template '%s' was created successfully.", name)

	return nil
}

func (c *PolicyApp) getGeneratorConfig(tmpInfo templateInfo) (generators.Config, error) {
	genConfig := generators.Config{}
	var err error
	if tmpInfo.kind == "policy" {
		return genConfig, nil
	}

	genConfig.Server, err = c.getDefaultServer()
	if err != nil {
		return genConfig, err
	}
	respServer := ""

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("server (%s)", genConfig.Server), &respServer,
	).Do()

	respServer = strings.TrimSpace(respServer)

	if respServer != "" {
		genConfig.Server = respServer
	}

	genConfig.User = c.getDefaultUser(genConfig.Server)
	respUser := ""

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("user (%s)", genConfig.User), &respUser,
	).Do()

	respUser = strings.TrimSpace(respUser)

	if respUser != "" {
		genConfig.User = respUser
	}

	respTokenName := ""
	genConfig.Token = c.getDefaultTokenName(genConfig.Server)

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("secret name (%s)", genConfig.Token), &respTokenName,
	).Do()

	respTokenName = strings.TrimSpace(respTokenName)

	if respTokenName != "" {
		genConfig.Token = respTokenName
	}

	respRepo := ""
	c.UI.Normal().Compact().WithAskString(
		"org/repo", &respRepo,
	).Do()

	if !strings.Contains(respRepo, "/") {
		return genConfig, errors.New("repo must be in the format 'org/repo'")
	}
	genConfig.Repo = strings.TrimSpace(respRepo)

	return genConfig, nil
}

func (c *PolicyApp) getDefaultTokenName(server string) string {
	if token, ok := c.Configuration.TokenDefaults[server]; ok {
		return token
	}

	if token, ok := c.Configuration.TokenDefaults[c.Configuration.DefaultDomain]; ok {
		return token
	}

	return ""
}

func (c *PolicyApp) getDefaultUser(server string) string {
	if s, err := c.Configuration.CredentialsStore.Get(server); err != nil {
		return s.Username
	}
	return ""
}

func (c *PolicyApp) getDefaultServer() (string, error) {

	if c.Configuration.DefaultDomain != "" {
		return c.Configuration.DefaultDomain, nil
	}

	servers, err := c.Configuration.CredentialsStore.GetAll()
	if err != nil {
		return "", err
	}
	var serverlist []string
	for name := range servers {
		serverlist = append(serverlist, name)
	}

	return c.buildTable("server", serverlist), nil
}

func (c *PolicyApp) buildTable(name string, items []string) string {
	sort.Strings(items)

	allowedValues := make([]int, len(items))
	table := c.UI.Normal().WithTable("#", name)
	for i, item := range items {
		table.WithTableRow(strconv.Itoa(i+1), item)
		allowedValues[i] = i + 1
	}

	table.Do()
	var response int64
	c.UI.Normal().Compact().WithAskInt(fmt.Sprintf("Select %s#", name), &response, allowedValues...).Do()

	return items[response-1]
}
