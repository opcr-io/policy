package app

import (
	"os"
	dirpath "path"
	"path/filepath"
	"strings"
	"text/template"

	git "github.com/go-git/go-git/v5/config"
	"github.com/opcr-io/policy/templates"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	gitDir        string = ".git"
	gitConfig     string = "config"
	gitOrigin     string = "origin"
	gitIgnore     string = ".gitignore"
	githubDir     string = ".github"
	githubConfig  string = "config.yaml"
	workflowsDir  string = "workflows"
	workflowsFile string = "build-release-policy.yaml"
	srcDir        string = "src"
	policiesDir   string = "policies"
	manifestFile  string = ".manifest"
	regoFile      string = "hello.rego"
	makeFile      string = "Makefile"
	readmeFile    string = "README.md"
)

// Init
// path is rootpath of project
func (c *PolicyApp) Init(path, user, server, repo, scc, secret string, overwrite bool) error {
	defer c.Cancel()

	if !strings.EqualFold(scc, "github") {
		return errors.Errorf("not supported source code provider '%s'", scc)
	}

	if exist, _ := dirExist(path); !exist {
		return errors.Errorf("root path not a directory '%s'", path)
	}

	if err := isGitRepo(path); err != nil {
		return err
	}

	if err := hasGitRemote(path); err != nil {
		c.UI.Exclamation().Msg("git repo does not contain a remote")
	}

	fns := []func() error{
		writeGitIgnore(path, overwrite),
		writeGithubConfig(path, overwrite, user, server, repo),
		writeGithubWorkflow(path, overwrite, secret),
		writeManifest(path, overwrite),
		writeRegoSourceFile(path, overwrite),
		writeMakefile(path, overwrite),
		writeReadMe(path, overwrite),
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func isGitRepo(path string) error {
	if exist, _ := dirExist(filepath.Join(path, gitDir)); !exist {
		return errors.Errorf("root path does not contain .git directory '%s'", path)
	}
	if exist, _ := fileExist(filepath.Join(path, gitDir, gitConfig)); !exist {
		return errors.Errorf(".git directory does not contain config file '%s'", path)
	}
	return nil
}

func hasGitRemote(path string) error {
	filePath := filepath.Join(path, gitDir, gitConfig)
	r, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "opening file '%s'", filePath)
	}

	gitConfig, err := git.ReadConfig(r)
	if err != nil {
		return errors.Wrapf(err, "reading file '%s'", filePath)
	}

	if len(gitConfig.Remotes) == 0 {
		return errors.Errorf("no remotes configured")
	}

	if _, ok := gitConfig.Remotes[gitOrigin]; !ok {
		return errors.Errorf("no origin remote configured")
	}

	return nil
}

func writeGitIgnore(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path)
		return writeTemplate(dirPath, gitIgnore, "github/gitignore.tmpl", overwrite)
	}
}

func writeGithubConfig(path string, overwrite bool, params ...string) func() error {
	return func() error {
		var (
			user   = params[0]
			server = params[1]
			repo   = params[2]
		)
		dirPath := dirpath.Join(path, githubDir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return errors.Wrapf(err, "create directory '%s'", dirPath)
		}

		filePath := filepath.Join(dirPath, githubConfig)

		exist, _ := fileExist(filePath)
		if exist && !overwrite {
			return nil
		}

		cfg := struct {
			Server   string `yaml:"server"`
			Username string `yaml:"username"`
			Repo     string `yaml:"repo"`
		}{
			Username: user,
			Server:   server,
			Repo:     repo,
		}

		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrapf(err, "open file '%s'", filePath)
		}
		defer f.Close()
		enc := yaml.NewEncoder(f)
		if err := enc.Encode(cfg); err != nil {
			return errors.Wrapf(err, "encode file '%s'", filePath)
		}

		return nil
	}
}

func writeGithubWorkflow(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path, githubDir, workflowsDir)
		paramss := struct {
			PushKey string
		}{
			PushKey: params[0],
		}
		return writeTemplate(dirPath, workflowsFile, "github/build-release-policy.tmpl", overwrite, paramss)
	}
}

func writeManifest(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path, srcDir)
		return writeTemplate(dirPath, manifestFile, "opa/manifest.tmpl", overwrite)
	}
}

func writeRegoSourceFile(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path, srcDir, policiesDir)
		return writeTemplate(dirPath, regoFile, "opa/hello-rego.tmpl", overwrite)
	}
}

func writeMakefile(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path)
		return writeTemplate(dirPath, makeFile, "general/makefile.tmpl", overwrite)
	}
}

func writeReadMe(path string, overwrite bool, params ...string) func() error {
	return func() error {
		dirPath := dirpath.Join(path)
		return writeTemplate(dirPath, readmeFile, "general/readme.tmpl", overwrite)
	}
}

func fileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func dirExist(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat directory '%s'", path)
	}
}

func writeTemplate(dirPath, fileName, templateName string, overwrite bool, params ...interface{}) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return errors.Wrapf(err, "create directory '%s'", dirPath)
	}

	filePath := filepath.Join(dirPath, fileName)

	exist, _ := fileExist(filePath)
	if exist && !overwrite {
		return nil
	}

	w, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "open file '%s'", filePath)
	}
	defer w.Close()

	aFS := templates.Assets()
	name := filepath.Base(templateName)
	t, err := template.New(name).ParseFS(aFS, templateName)
	if err != nil {
		return err
	}

	var data interface{} = nil
	if len(params) == 1 {
		data = params[0]
	}

	if err := t.Execute(w, data); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
