package generators

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/pkg/errors"
)

const (
	makeFile   string = "Makefile"
	readmeFile string = "README.md"
)

type General struct {
	path       string
	policyName string
	ui         *clui.UI
}

func NewGeneral(path, policyName string, ui *clui.UI) *General {
	dirPath := filepath.Join(path)
	return &General{
		path:       dirPath,
		policyName: policyName,
		ui:         ui,
	}
}

func (g *General) GenerateFilesContent() (GeneratedFilesContent, error) {
	result := make(GeneratedFilesContent)
	var err error

	// interpolate makefile
	makefilePath := filepath.Join(g.path, makeFile)
	result[makefilePath], err = InterpolateTemplate("general/makefile.tmpl", nil)
	if err != nil {
		return result, err
	}

	// interpolate readme
	param := struct {
		PolicyName string
	}{PolicyName: g.policyName}
	readmeFilePath := filepath.Join(g.path, readmeFile)
	result[readmeFilePath], err = InterpolateTemplate("general/readme.tmpl", param)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (g *General) Generate(overwrite bool) error {
	if err := os.MkdirAll(g.path, 0755); err != nil {
		return errors.Wrapf(err, "create directory '%s'", g.path)
	}

	if err := g.writeMakeFile(overwrite); err != nil {
		return err
	}

	return g.writeReadMe(overwrite)
}

func (g *General) writeReadMe(overwrite bool) error {
	filePath := filepath.Join(g.path, readmeFile)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	param := struct {
		PolicyName string
	}{PolicyName: g.policyName}

	return WriteTemplateToFile(filePath, "general/readme.tmpl", param)
}

func (g *General) writeMakeFile(overwrite bool) error {
	filePath := filepath.Join(g.path, makeFile)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}
	return WriteTemplateToFile(filePath, "general/makefile.tmpl", nil)
}
