package app

import (
	"sort"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/policytemplates"
	"github.com/pkg/errors"
)

type templateInfo struct {
	name        string
	kind        string
	description string
}

func (c *PolicyApp) TemplatesList() error {
	defer c.Cancel()

	templateInfos, err := c.listTemplates()
	if err != nil {
		return err
	}

	sort.Slice(templateInfos, func(i, j int) bool {
		return templateInfos[i].name < templateInfos[j].name
	})

	table := c.UI.Normal().WithTable("Name", "Kind", "Description")

	for _, tmplInfo := range templateInfos {
		if tmplInfo.kind == "" {
			continue
		}
		table.WithTableRow(tmplInfo.name, tmplInfo.kind, tmplInfo.description)
	}

	table.WithTableNoAutoWrapText().Do()

	return nil
}

func (c *PolicyApp) listTemplates() ([]templateInfo, error) {

	templateInfos, err := c.getTemplates(
		c.Configuration.ContentTemplates.Server,
		c.Configuration.ContentTemplates.Organization,
		c.Configuration.ContentTemplates.Tag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	ciTemplates, err := c.getTemplates(
		c.Configuration.CITemplates.Server,
		c.Configuration.CITemplates.Organization,
		c.Configuration.CITemplates.Tag,
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	return append(templateInfos, ciTemplates...), nil

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
