package app

import (
	"fmt"
	"sort"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/scc-lib/generators"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/policytemplates"
	"github.com/pkg/errors"
)

type templateInfo struct {
	name        string
	kind        string
	description string
}

func (c *PolicyApp) Templates(name, output string, list, overwrite bool) error {

	if name == "" {
		return errors.New("template name is required")
	}

	prog := c.UI.Progressf("Processing template '%s'", name)
	prog.Start()
	policyTemplatesCfg := policytemplates.Config{
		Server:     c.Configuration.CITemplates.Server,
		PolicyRoot: c.Configuration.PoliciesRoot(),
	}

	ciTemplates := policytemplates.NewOCI(c.Context, c.Logger, c.TransportWithTrustedCAs(), policyTemplatesCfg)

	templateFs, err := ciTemplates.Load(
		fmt.Sprintf("%s/%s:%s",
			c.Configuration.ContentTemplates.Organization,
			name,
			c.Configuration.ContentTemplates.Tag))
	if err != nil {
		return errors.Wrapf(err, "failed to load '%s' ci template", name)
	}

	generator, err := generators.NewGenerator(
		&generators.Config{},
		c.Logger,
		templateFs,
	)

	if err != nil {
		return errors.Wrap(err, "failed to initialize generator")
	}
	err = generator.Generate(output, overwrite)
	if err != nil {
		return errors.Wrap(err, "failed to generate policy")
	}

	prog.Stop()

	c.UI.Normal().Msgf("The template '%s' was created successfully.", name)

	return nil
}

func (c *PolicyApp) TemplatesList() error {

	templateInfos, err := c.getTemplates(
		c.Configuration.ContentTemplates.Server,
		c.Configuration.ContentTemplates.Organization,
		c.Configuration.ContentTemplates.Tag)
	if err != nil {
		return errors.Wrap(err, "failed to list templates")
	}

	ciTemplates, err := c.getTemplates(
		c.Configuration.CITemplates.Server,
		c.Configuration.CITemplates.Organization,
		c.Configuration.CITemplates.Tag,
	)

	if err != nil {
		return errors.Wrap(err, "failed to list templates")
	}

	templateInfos = append(templateInfos, ciTemplates...)

	sort.Slice(templateInfos, func(i, j int) bool {
		return templateInfos[i].name < templateInfos[j].name
	})

	table := c.UI.Normal().WithTable("Name", "Kind", "Description")

	for _, tmplInfo := range templateInfos {
		table.WithTableRow(tmplInfo.name, tmplInfo.kind, tmplInfo.description)
	}

	table.WithTableNoAutoWrapText().Do()

	return nil
}

func (c *PolicyApp) getTemplates(server, org, tag string) ([]templateInfo, error) {
	var tmplInfo []templateInfo

	policyTemplatesCfg := policytemplates.Config{
		Server:     server,
		PolicyRoot: c.Configuration.PoliciesRoot(),
	}
	policyTmpl := policytemplates.NewOCI(
		c.Context,
		c.Logger,
		c.TransportWithTrustedCAs(),
		policyTemplatesCfg)

	tmplRepo, err := policyTmpl.ListRepos(org, tag)

	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}
	for repo, tag := range tmplRepo {
		vendor, description := getDetails(tag.Annotations)
		tmplInfo = append(tmplInfo, templateInfo{
			name:        repo,
			kind:        vendor,
			description: description,
		})
	}

	return tmplInfo, nil
}

func getDetails(annotations []*api.RegistryRepoAnnotation) (kind, description string) {
	for _, annotation := range annotations {
		if annotation == nil {
			continue
		}
		if annotation.Key == extendedregistry.AnnotationPolicyRegistryTemplateKind {
			kind = annotation.Value
		}
		if annotation.Key == extendedregistry.AnnotationImageDescription {
			description = annotation.Value
		}
	}
	return
}
