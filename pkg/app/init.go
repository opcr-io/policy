package app

import (
	"fmt"
	"os"
	dirpath "path"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5/config"
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
	buildDir      string = "build"
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

	if strings.ToLower(scc) != "github" {
		return errors.Errorf("not supported source code provider '%s'", scc)
	}

	if exist, _ := dirExist(path); !exist {
		return errors.Errorf("root path not a directory '%s'", path)
	}

	if err := isGitRepo(path); err != nil {
		return err
	}

	if err := hasGitRemote(path); err != nil {
		return err
	}

	if err := hasGitIgnoreFile(path, overwrite); err != nil {
		return err
	}

	if err := hasGithubConfig(path, user, server, repo, overwrite); err != nil {
		return err
	}

	if err := hasGithubWorkflow(path, secret, overwrite); err != nil {
		return err
	}

	if err := hasManifest(path, overwrite); err != nil {
		return err
	}

	if err := hasRegoSourceFile(path, overwrite); err != nil {
		return err
	}

	if err := hasMakefile(path, overwrite); err != nil {
		return err
	}

	if err := hasReadMe(path, overwrite); err != nil {
		return err
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

	if _, ok := gitConfig.Remotes["origin"]; !ok {
		return errors.Errorf("no origin remote configured")
	}

	return nil
}

func hasGitIgnoreFile(path string, overwrite bool) error {
	dirPath := dirpath.Join(path)
	return writeTemplate(dirPath, gitIgnore, gitignoreTemplate, overwrite)
}

func hasGithubConfig(path, user, server, repo string, overwrite bool) error {
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

func hasGithubWorkflow(path, secret string, overwrite bool) error {
	dirPath := dirpath.Join(path, githubDir, workflowsDir)
	return writeTemplate(dirPath, workflowsFile, workflowTemplate, overwrite, secret, secret)
}

func hasManifest(path string, overwrite bool) error {
	dirPath := dirpath.Join(path, srcDir)
	return writeTemplate(dirPath, manifestFile, manifestTemplate, overwrite)
}

func hasRegoSourceFile(path string, overwrite bool) error {
	dirPath := dirpath.Join(path, srcDir)
	return writeTemplate(dirPath, regoFile, regoTemplate, overwrite)
}

func hasMakefile(path string, overwrite bool) error {
	dirPath := dirpath.Join(path)
	return writeTemplate(dirPath, makeFile, makeFileTemplate, overwrite)
}

func hasReadMe(path string, overwrite bool) error {
	dirPath := dirpath.Join(path)
	return writeTemplate(dirPath, readmeFile, readmeTemplate, overwrite)
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
		return false, errors.Wrapf(err, "failed to stat directory '%s', path")
	}
}

func writeTemplate(dirPath, fileName, template string, overwrite bool, params ...interface{}) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return errors.Wrapf(err, "create directory '%s'", dirPath)
	}

	filePath := filepath.Join(dirPath, fileName)

	exist, _ := fileExist(filePath)
	if exist && !overwrite {
		return nil
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "open file '%s'", filePath)
	}
	defer f.Close()

	fmt.Fprintf(f, template, params...)

	return nil
}

const gitignoreTemplate string = `.DS_Store
bundle.tar.gz
build/
`

const workflowTemplate string = `name: build-release

on:
  workflow_dispatch:
  push:
    tags:
    - '*'

jobs:
  release_policy:
    runs-on: ubuntu-latest
    name: build

    steps:
    - uses: actions/checkout@v2

    - name: Read config
      id: config
      uses: CumulusDS/get-yaml-paths-action@v0.1.1
      with:
        file: .github/config.yaml
        username: username
        repo: repo
        server: server

    - name: List Sver Tags
      uses: aserto-dev/sver-action@v0.0.14
      id: "sver"
      with:
        docker_image: ${{ steps.config.outputs.repo }}
        docker_registry: ${{ steps.config.outputs.server }}
        docker_username: ${{ steps.config.outputs.username }}
        docker_password: ${{ secrets.%s }}

    - name: Calculate image tags
      id: "tags"
      run: |
        while read -r tag; do
          tags="$tags ${{ steps.config.outputs.repo }}:$tag"
        done < <(echo "${{ steps.sver.outputs.version }}")

        echo ::set-output name=target_tags::$tags

    - name: Policy Login
      id: policy-login
      uses: opcr-io/policy-login-action@v2
      with:
        username: ${{ steps.config.outputs.username }}
        password: ${{ secrets.%s }}
        server: ${{ steps.config.outputs.server }}

    - name: Policy Build
      id: policy-build
      uses: opcr-io/policy-build-action@v2
      with:
        src: src
        tag: ${{ steps.config.outputs.repo }}
        revision: "$GITHUB_SHA"
      env:
        POLICY_DEFAULT_DOMAIN: ${{ steps.config.outputs.server }}

    - name: Policy Tag
      id: policy-tag
      uses: opcr-io/policy-tag-action@v2
      with:
        source_tag: ${{ steps.config.outputs.repo }}
        target_tags: ${{ steps.tags.outputs.target_tags }}
      env:
        POLICY_DEFAULT_DOMAIN: ${{ steps.config.outputs.server }}

    - name: Policy Push
      id: policy-push
      uses: opcr-io/policy-push-action@v2
      with:
        tags: ${{ steps.tags.outputs.target_tags }}
      env:
        POLICY_DEFAULT_DOMAIN: ${{ steps.config.outputs.server }}

    - name: Policy Logout
      id: policy-logout
      uses: opcr-io/policy-logout-action@v2
      with:
        server: ${{ steps.config.outputs.server }}
`

const manifestTemplate string = `{
    "roots": [""],
    "metadata": {
      "required_builtins": {
          "builtin1": [
              {
                  "name": "dir.user",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "dir.manager_of",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "dir.management_chain",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "res.get",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "dir.identity",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              }
          ],
          "builtin2": [
              {
                  "name": "dir.is_manager_of",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          },
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "boolean"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "dir.works_for",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          },
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "boolean"
                      },
                      "type": "function"
                  }
              },
              {
                  "name": "dir.is_same_user",
                  "decl": {
                      "args": [
                          {
                              "type": "string"
                          },
                          {
                              "type": "string"
                          }
                      ],
                      "result": {
                          "type": "boolean"
                      },
                      "type": "function"
                  }
              }
          ],
          "builtinDyn": [
              {
                  "name": "res.list",
                  "decl": {
                      "result": {
                          "type": "any"
                      },
                      "type": "function"
                  }
              }
          ]
      }
  }
}
`
const regoTemplate string = `package policies.hello

# default to a "closed" system, 
# only grant access when explicitly granted

default allowed = false
default visible = false
default enabled = false

allowed {
    input.role == "web-admin"
}

enabled {
    visible
}

visible {
    input.app == "web-console"
}
`
const makeFileTemplate string = `SHELL 	   := $(shell which bash)

NO_COLOR   :=\033[0m
OK_COLOR   :=\033[32;01m
ERR_COLOR  :=\033[31;01m
WARN_COLOR :=\033[36;01m
ATTN_COLOR :=\033[33;01m

.PHONY: all 
all: login build tag push logout

.PHONY: login
login:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: build
build:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: tag
tag:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: push
push:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: logout
logout:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
`

const readmeTemplate string = ``
