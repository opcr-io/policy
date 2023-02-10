package app

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/opcr-io/policy/templates"
	"github.com/pkg/errors"
)

type templateInfo struct {
	name        string
	kind        string
	description string
}

func (c *PolicyApp) TemplatesList() error {
	defer c.Cancel()
	p := c.UI.Progress("Fetching templates")
	p.Start()

	templateInfos, err := c.listTemplates()
	if err != nil {
		return err
	}
	p.Stop()

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

	var list []templateInfo
	err := fs.WalkDir(templates.Assets(), ".", func(path string, d fs.DirEntry, err error) error {
		if d.Name() != "." {
			if strings.Contains(path, "github") || strings.Contains(path, "gitlab") {
				if d.IsDir() {
					list = append(list, templateInfo{name: d.Name(), kind: "cicd", description: fmt.Sprintf("%s template", d.Name())})
				}
			} else {
				if d.IsDir() {
					list = append(list, templateInfo{name: d.Name(), kind: "policy", description: fmt.Sprintf("%s template", d.Name())})
				}
			}
		}
		return nil
	})

	// templateInfos, err := c.getTemplates(
	// 	c.Configuration.ContentTemplates.Server,
	// 	c.Configuration.ContentTemplates.Organization,
	// 	c.Configuration.ContentTemplates.Tag)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to list templates")
	// }

	// ciTemplates, err := c.getTemplates(
	// 	c.Configuration.CITemplates.Server,
	// 	c.Configuration.CITemplates.Organization,
	// 	c.Configuration.CITemplates.Tag,
	// )

	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	return list, nil

}

// func (c *PolicyApp) getTemplates(server, org, tag string) ([]templateInfo, error) {
// 	var tmplInfo []templateInfo

// 	policyTemplatesCfg := policytemplates.Config{
// 		Server:     server,
// 		PolicyRoot: c.Configuration.PoliciesRoot(),
// 	}
// 	policyTmpl := policytemplates.NewOCI(
// 		c.Context,
// 		c.Logger,
// 		c.TransportWithTrustedCAs(),
// 		policyTemplatesCfg)

// 	tmplRepo, err := policyTmpl.ListRepos(org, tag)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to list templates")
// 	}
// 	for repo, tag := range tmplRepo {
// 		vendor, description := getDetails(tag.Annotations)
// 		tmplInfo = append(tmplInfo, templateInfo{
// 			name:        repo,
// 			kind:        vendor,
// 			description: description,
// 		})
// 	}

// 	return tmplInfo, nil
// }

// func getDetails(annotations []*api.RegistryRepoAnnotation) (kind, description string) {
// 	for _, annotation := range annotations {
// 		if annotation == nil {
// 			continue
// 		}
// 		if annotation.Key == extendedregistry.AnnotationPolicyRegistryTemplateKind {
// 			kind = annotation.Value
// 		}
// 		if annotation.Key == extendedregistry.AnnotationImageDescription {
// 			description = annotation.Value
// 		}
// 	}
// 	return
// }
