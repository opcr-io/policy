package policytemplates

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestOCIListRepos(t *testing.T) {
	assert := require.New(t)

	log := zerolog.New(os.Stdout)
	transport := &http.Transport{}
	ctx := context.Background()
	expectedRepo := "peoplefinder-rbac"

	ociTemplate := NewOCI(ctx, &log, transport, Config{
		Server:     "opcr.io",
		PolicyRoot: "",
	})

	templateRepos, err := ociTemplate.ListRepos("aserto-templates", "1")
	if err != nil {
		t.Fatal(err)
	}

	exists := false
	for repo := range templateRepos {
		if repo == expectedRepo {
			exists = true
			break
		}
	}
	assert.Truef(exists, "expected repo %s not found", expectedRepo)
}

func TestOCILoadRepo(t *testing.T) {
	assert := require.New(t)

	expectedFiles := map[string]bool{
		".":                       false,
		".gitignore":              false,
		"src":                     false,
		"src/policies":            false,
		"src/policies/hello.rego": false,
		"src/.manifest":           false,
		"README.md":               false,
	}

	log := zerolog.New(os.Stdout)
	transport := &http.Transport{}
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "ociload")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	ociTemplate := NewOCI(ctx, &log, transport, Config{
		Server:     "opcr.io",
		PolicyRoot: tmpDir,
	})

	bundleFS, err := ociTemplate.Load("aserto-content/policy-template:1.0.1")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.WalkDir(bundleFS, ".", func(bundlePath string, d fs.DirEntry, err error) error {
		if _, ok := expectedFiles[bundlePath]; !ok {
			assert.Failf("file not expected", "unexpected file '%s'", bundlePath)
		} else {
			expectedFiles[bundlePath] = true
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	for filePath, exists := range expectedFiles {
		if !exists {
			assert.Failf("file missing", "file '%s' not found", filePath)
		}
	}
}
