package generators

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/pkg/errors"
)

const (
	githubDir     string = ".github"
	githubConfig  string = "config.yaml"
	workflowsDir  string = "workflows"
	workflowsFile string = "build-release-policy.yaml"
)

type githubCIParams struct {
	Server   string
	Username string
	Repo     string
}

type Github struct {
	path   string
	server string
	repo   string
	token  string
	user   string
	ui     *clui.UI
}

func NewGithub(scc *SCC) *Github {
	dirPath := filepath.Join(scc.Path)
	return &Github{
		path:   dirPath,
		server: scc.Server,
		user:   scc.User,
		repo:   scc.Repo,
		token:  scc.Token,
		ui:     scc.UI,
	}
}

func (g *Github) Generate(overwrite bool) error {
	gitDir := filepath.Join(g.path, githubDir)
	gitWorkflowDir := filepath.Join(gitDir, workflowsDir)
	err := os.MkdirAll(gitWorkflowDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "create directory '%s'", gitWorkflowDir)
	}

	err = g.writeGitIgnore(overwrite)
	if err != nil {
		return err
	}

	err = g.writeGithubConfig(gitDir, overwrite)
	if err != nil {
		return err
	}

	return g.writeGithubWorkflow(gitWorkflowDir, overwrite)
}

func (g *Github) GenerateFilesContent() (GeneratedFilesContent, error) {
	result := make(GeneratedFilesContent)
	var err error

	// interpolate gitignore
	gitIgnoreFile := filepath.Join(g.path, gitIgnore)
	result[gitIgnoreFile], err = InterpolateTemplate("github/gitignore.tmpl", nil)
	if err != nil {
		return result, err
	}

	// interpolate config.yml
	cfg := githubCIParams{
		Username: g.user,
		Server:   g.server,
		Repo:     g.repo,
	}

	gitDir := filepath.Join(g.path, githubDir)
	configFile := filepath.Join(gitDir, githubConfig)
	result[configFile], err = InterpolateTemplate("github/config.tmpl", cfg)
	if err != nil {
		return result, err
	}

	// interpolate workflow
	param := struct {
		PushKey string
	}{PushKey: g.token}
	gitWorkflowDir := filepath.Join(gitDir, workflowsDir)
	workflowFile := filepath.Join(gitWorkflowDir, workflowsFile)
	result[workflowFile], err = InterpolateTemplate("github/build-release-policy.tmpl", param)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (g *Github) writeGitIgnore(overwrite bool) error {
	filePath := filepath.Join(g.path, gitIgnore)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	return WriteTemplateToFile(filePath, "github/gitignore.tmpl", nil)
}

func (g *Github) writeGithubConfig(gitDir string, overwrite bool) error {
	filePath := filepath.Join(gitDir, githubConfig)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	cfg := githubCIParams{
		Username: g.user,
		Server:   g.server,
		Repo:     g.repo,
	}

	return WriteTemplateToFile(filePath, "github/config.tmpl", cfg)
}

func (g *Github) writeGithubWorkflow(gitWorkflowDir string, overwrite bool) error {
	filePath := filepath.Join(gitWorkflowDir, workflowsFile)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	param := struct {
		PushKey string
	}{PushKey: g.token}

	return WriteTemplateToFile(filePath, "github/build-release-policy.tmpl", param)
}
