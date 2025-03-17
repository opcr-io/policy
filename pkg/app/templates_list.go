package app

import (
	"io/fs"
	"os"
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

var templatesDescription = map[string]string{
	"github":          "GitHub policy CI/CD template.",
	"gitlab":          "GitLab policy CI/CD template.",
	"policy-template": "Minimal policy template.",
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
		if d.Name() != "." && !strings.Contains(path, string(os.PathSeparator)) {
			if !d.IsDir() {
				return nil
			}

			if strings.Contains(path, "github") || strings.Contains(path, "gitlab") {
				list = append(list, templateInfo{name: d.Name(), kind: "cicd", description: getDescription(d.Name())})

				return nil
			}

			list = append(list, templateInfo{name: d.Name(), kind: "policy", description: getDescription(d.Name())})
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list templates")
	}

	return list, nil
}

func getDescription(folderName string) string {
	val, ok := templatesDescription[folderName]
	if !ok {
		return "no description available"
	}

	return val
}
