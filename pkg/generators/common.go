package generators

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/opcr-io/policy/templates"
	"github.com/pkg/errors"
)

const (
	gitIgnore string = ".gitignore"
	gitDir    string = ".git"
	gitConfig string = "config"
)

type SCC struct {
	Path   string
	Server string
	Repo   string
	Token  string
	User   string
	UI     *clui.UI
}

func IsGitRepo(path string) error {
	if exist, _ := DirExist(filepath.Join(path, gitDir)); !exist {
		return errors.Errorf("root path does not contain .git directory '%s'", path)
	}
	if exist, _ := FileExist(filepath.Join(path, gitDir, gitConfig)); !exist {
		return errors.Errorf(".git directory does not contain config file '%s'", path)
	}
	return nil
}

func FileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func DirExist(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat directory '%s'", path)
	}
}

func WriteTemplateToFile(filePath, templateName string, params interface{}) error {
	w, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "open file '%s'", filePath)
	}
	defer w.Close()

	content, err := InterpolateTemplate(templateName, params)
	if err != nil {
		return err
	}

	_, err = w.WriteString(content)
	if err != nil {
		return err
	}

	return w.Close()
}

func InterpolateTemplate(templateName string, params interface{}) (string, error) {
	aFS := templates.Assets()
	name := filepath.Base(templateName)
	t, err := template.New(name).ParseFS(aFS, templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err := t.ExecuteTemplate(&buf, name, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}
