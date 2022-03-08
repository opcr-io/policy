package generators

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/pkg/errors"
)

const (
	gitlabCI string = ".gitlab-ci.yml"
)

type gitlabCIParams struct {
	Repo     string
	Server   string
	Username string
	PushKey  string
}

type Gitlab struct {
	path        string
	server      string
	repo        string
	pushKeyName string
	user        string
	ui          *clui.UI
}

func NewGitlab(scc *SCC) *Gitlab {
	dirPath := filepath.Join(scc.Path)
	pushKeyName := "$" + scc.Token
	return &Gitlab{
		path:        dirPath,
		server:      scc.Server,
		user:        scc.User,
		repo:        scc.Repo,
		pushKeyName: pushKeyName,
		ui:          scc.UI,
	}
}

func (g *Gitlab) Generate(overwrite bool) error {
	err := os.MkdirAll(g.path, 0755)
	if err != nil {
		return errors.Wrapf(err, "create directory '%s'", g.path)
	}

	err = g.writeGitIgnore(overwrite)
	if err != nil {
		return err
	}

	return g.writeGitlabCI(overwrite)
}

func (g *Gitlab) GenerateFilesContent() (GeneratedFilesContent, error) {
	result := make(GeneratedFilesContent)
	var err error

	// interpolate gitignore
	gitIgnoreFile := filepath.Join(g.path, gitIgnore)
	result[gitIgnoreFile], err = InterpolateTemplate("gitlab/gitignore.tmpl", nil)
	if err != nil {
		return result, err
	}

	// interpolate workflow
	param := gitlabCIParams{
		Repo:     g.repo,
		Server:   g.server,
		Username: g.user,
		PushKey:  g.pushKeyName,
	}
	workflowFile := filepath.Join(g.path, gitlabCI)
	result[workflowFile], err = InterpolateTemplate("gitlab/gitlab-ci.tmpl", param)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (g *Gitlab) writeGitlabCI(overwrite bool) error {
	filePath := filepath.Join(g.path, gitlabCI)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	param := gitlabCIParams{
		Repo:     g.repo,
		Server:   g.server,
		Username: g.user,
		PushKey:  g.pushKeyName,
	}

	return WriteTemplateToFile(filePath, "gitlab/gitlab-ci.tmpl", param)
}

func (g *Gitlab) writeGitIgnore(overwrite bool) error {
	filePath := filepath.Join(g.path, gitIgnore)

	exist, _ := FileExist(filePath)
	if exist && !overwrite {
		if g.ui != nil {
			g.ui.Exclamation().Msgf("file '%s' already exists, skipping", filePath)
		}
		return nil
	}

	return WriteTemplateToFile(filePath, "gitlab/gitignore.tmpl", nil)
}
