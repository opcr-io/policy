package generators

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"io/fs"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GeneratedFilesContent map[string]string

type Generator interface {
	GenerateFilesContent() (GeneratedFilesContent, error)
	Generate(pathToTemplates string, overwrite bool) error
}

type generatorImpl struct {
	cfg    *Config
	files  []string
	logger *zerolog.Logger
	dfs    fs.FS
}

func NewGenerator(cfg *Config, log *zerolog.Logger, dfs fs.FS) (Generator, error) {
	if log == nil {
		return nil, errors.New("no logger variable provided")
	}

	var files []string
	_ = fs.WalkDir(dfs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})
	return &generatorImpl{
		cfg:    cfg,
		files:  files,
		dfs:    dfs,
		logger: log,
	}, nil
}

func (c *generatorImpl) GenerateFilesContent() (GeneratedFilesContent, error) {
	result := make(GeneratedFilesContent)

	for _, file := range c.files {
		if !strings.Contains(file, ".tmpl") {
			content, err := fs.ReadFile(c.dfs, file)
			if err != nil {
				return result, err
			}
			result[file] = string(content)
			continue
		}
		content, err := c.interpolateTemplate(file)
		if err != nil {
			return result, err
		}
		fileName := strings.TrimSuffix(file, ".tmpl")
		result[fileName] = content
	}

	return result, nil
}

func (c *generatorImpl) Generate(pathToTemplates string, overwrite bool) error {
	for _, file := range c.files {
		fileName := filepath.Join(pathToTemplates, strings.TrimSuffix(file, ".tmpl"))

		// check if file exists
		exist, err := FileExist(fileName)
		if err != nil {
			return err
		}
		if exist && !overwrite {
			c.logger.Debug().Msgf("file '%s' already exists, skipping", fileName)
			continue
		}

		// create directories path
		baseDir := filepath.Dir(fileName)
		err = os.MkdirAll(baseDir, 0755)
		if err != nil {
			return errors.Wrapf(err, "create directory '%s'", baseDir)
		}

		var content string
		if strings.Contains(file, ".tmpl") {
			content, err = c.interpolateTemplate(file)
			if err != nil {
				return err
			}
		} else {
			cnt, err := fs.ReadFile(c.dfs, file)
			if err != nil {
				return err
			}
			content = string(cnt)
		}

		err = c.writeContentToFile(fileName, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *generatorImpl) writeContentToFile(filePath, content string) error {
	w, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "open file '%s'", filePath)
	}
	defer w.Close()

	_, err = w.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func (c *generatorImpl) interpolateTemplate(templateName string) (string, error) {
	funcs := template.FuncMap{
		"server": func() string {
			return c.cfg.Server
		},
		"repo": func() string {
			return c.cfg.Repo
		},
		"username": func() string {
			return c.cfg.User
		},
		"pushkey": func() string {
			return c.cfg.Token
		},
	}

	parsedTemplate, err := template.New(filepath.Base(templateName)).
		Funcs(sprig.TxtFuncMap()).
		Funcs(funcs).
		ParseFS(c.dfs, templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err := parsedTemplate.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}
