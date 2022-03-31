package app

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/scc-lib/generators"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/policytemplates"
	"github.com/pkg/errors"
)

func (c *PolicyApp) New(name, output string, list, overwrite bool) error {
	if list {
		return c.listTemplates(name)
	}
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

func (c *PolicyApp) listTemplates(name string) error {
	policyTemplatesCfg := policytemplates.Config{
		Server:     c.Configuration.ContentTemplates.Server,
		PolicyRoot: c.Configuration.PoliciesRoot(),
	}
	policyTmpl := policytemplates.NewOCI(
		c.Context,
		c.Logger,
		c.TransportWithTrustedCAs(),
		policyTemplatesCfg)

	tmplRepo, err := policyTmpl.ListRepos(
		c.Configuration.ContentTemplates.Organization,
		c.Configuration.ContentTemplates.Tag)

	if err != nil {
		return errors.Wrap(err, "failed to list templates")
	}
	table := c.UI.Normal().WithTable("Name", "Vendor", "Description")

	for repo, tag := range tmplRepo {
		if !strings.Contains(repo, name) {
			continue
		}

		vendor, description := getDetails(tag.Annotations)
		table.WithTableRow(repo, vendor, description)
	}
	table.WithTableNoAutoWrapText().Do()

	return nil
}

func getDetails(annotations []*api.RegistryRepoAnnotation) (vendor, description string) {
	for _, annotation := range annotations {
		if annotation == nil {
			continue
		}
		if annotation.Key == extendedregistry.AnnotationImageVendor {
			vendor = annotation.Value
		}
		if annotation.Key == extendedregistry.AnnotationImageDescription {
			description = annotation.Value
		}
	}
	return
}
