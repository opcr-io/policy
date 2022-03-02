package generators

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/pkg/errors"
)

const (
	srcDir       string = "src"
	policiesDir  string = "policies"
	manifestFile string = ".manifest"
	regoFile     string = "hello.rego"
)

type Opa struct {
	path string
	ui   *clui.UI
}

func NewOpa(path string, ui *clui.UI) *Opa {
	dirPath := filepath.Join(path)
	return &Opa{
		path: dirPath,
		ui:   ui,
	}
}

func (o *Opa) GenerateFilesContent() (GeneratedFilesContent, error) {
	result := make(GeneratedFilesContent)

	// interpolate manifest
	content, err := InterpolateTemplate("opa/manifest.tmpl", nil)
	if err != nil {
		return result, err
	}
	manifestFilePath := filepath.Join(o.path, srcDir, manifestFile)
	result[manifestFilePath] = content

	// interpolate rego file
	content, err = InterpolateTemplate("opa/hello-rego.tmpl", nil)
	if err != nil {
		return result, err
	}
	regoFilePath := filepath.Join(o.path, srcDir, policiesDir, regoFile)
	result[regoFilePath] = content

	return result, nil
}

func (o *Opa) Generate(overwrite bool) error {
	err := os.MkdirAll(o.path, 0755)
	if err != nil {
		return errors.Wrapf(err, "create directory '%s'", o.path)
	}

	policiesDirPath := filepath.Join(o.path, srcDir, policiesDir)
	err = os.MkdirAll(policiesDirPath, 0755)
	if err != nil {
		return errors.Wrapf(err, "create directory '%s'", policiesDirPath)
	}

	err = o.writeManifest(overwrite)
	if err != nil {
		return err
	}

	return o.writeRegoFile(policiesDirPath, overwrite)
}

func (o *Opa) writeManifest(overwrite bool) error {
	filePath := filepath.Join(o.path, srcDir, manifestFile)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if o.ui != nil {
			o.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}
	return WriteTemplateToFile(filePath, "opa/manifest.tmpl", nil)
}

func (o *Opa) writeRegoFile(policiesDir string, overwrite bool) error {
	filePath := filepath.Join(policiesDir, regoFile)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if o.ui != nil {
			o.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}
	return WriteTemplateToFile(filePath, "opa/hello-rego.tmpl", nil)
}
