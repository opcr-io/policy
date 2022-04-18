package app

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/aserto-dev/scc-lib/generators"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/policytemplates"
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

	var templateFs fs.FS

	switch tmplInfo.kind {
	case extendedregistry.TemplateTypeCICD:
		ciTemplateCfg := policytemplates.Config{
			Server:     c.Configuration.CITemplates.Server,
			PolicyRoot: c.Configuration.PoliciesRoot(),
		}
		ciTemplates := policytemplates.NewOCI(c.Context, c.Logger, c.TransportWithTrustedCAs(), ciTemplateCfg)
		templateFs, err = ciTemplates.Load(
			fmt.Sprintf("%s/%s:%s",
				c.Configuration.CITemplates.Organization,
				name,
				c.Configuration.CITemplates.Tag),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to load '%s' template", name)
		}
	case extendedregistry.TemplateTypePolicy:
		contentTemplateCfg := policytemplates.Config{
			Server:     c.Configuration.ContentTemplates.Server,
			PolicyRoot: c.Configuration.PoliciesRoot(),
		}
		ciTemplates := policytemplates.NewOCI(c.Context, c.Logger, c.TransportWithTrustedCAs(), contentTemplateCfg)
		templateFs, err = ciTemplates.Load(
			fmt.Sprintf("%s/%s:%s",
				c.Configuration.ContentTemplates.Organization,
				name,
				c.Configuration.ContentTemplates.Tag),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to load '%s' template", name)
		}
	default:
		return errors.Errorf("template '%s' has an unknown kind", name)
	}

	prog.Stop()

	generatorCfg, err := c.getGeneratorConfig(tmplInfo)
	if err != nil {
		return err
	}

	prog.ChangeMessage("Generating files")
	prog.Start()

	generator, err := generators.NewGenerator(
		&generatorCfg,
		c.Logger,
		templateFs,
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
	if tmpInfo.kind == extendedregistry.TemplateTypePolicy {
		return genConfig, nil
	}

	genConfig.Server = c.getDefaultServer()
	respServer := ""

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("server: (%s)", genConfig.Server), &respServer,
	).Do()

	respServer = strings.TrimSpace(respServer)

	if respServer != "" {
		genConfig.Server = respServer
	}

	genConfig.User = c.getDefaultUser(genConfig.Server)
	respUser := ""

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("user  : (%s)", genConfig.User), &respUser,
	).Do()

	respUser = strings.TrimSpace(respUser)

	if respUser != "" {
		genConfig.User = respUser
	}

	respTokenName := ""
	genConfig.Token = c.getDefaultTokenName(genConfig.Server)

	c.UI.Normal().Compact().WithAskString(
		fmt.Sprintf("secret name: (%s)", genConfig.Token), &respTokenName,
	).Do()

	respTokenName = strings.TrimSpace(respTokenName)

	if respTokenName != "" {
		genConfig.Token = respTokenName
	}

	respRepo := ""
	c.UI.Normal().Compact().WithAskString(
		"repo  : ()", &respRepo,
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
	if s, ok := c.Configuration.Servers[server]; ok {
		return s.Username
	}
	return ""
}

func (c *PolicyApp) getDefaultServer() string {
	if len(c.Configuration.Servers) == 0 {
		if c.Configuration.DefaultDomain != "" {
			return c.Configuration.DefaultDomain
		}
		return ""
	}

	servers := []string{}
	for name := range c.Configuration.Servers {
		servers = append(servers, name)
	}

	return c.buildTable("server", servers)
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
