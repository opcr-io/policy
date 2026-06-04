package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"policy": main,
	})
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:             "../../tests/cli",
		ContinueOnError: true,
		Setup: func(env *testscript.Env) error {
			policyBin, err := getPolicyBinPath()
			if err != nil {
				return err
			}

			_, file := filepath.Split(policyBin)

			binDir := filepath.Join(env.WorkDir, "bin")
			if err := os.MkdirAll(binDir, 0o750); err != nil {
				return err
			}

			if err := os.Symlink(policyBin, filepath.Join(binDir, file)); err != nil {
				return err
			}

			env.Vars = append(env.Vars, "PATH="+binDir+string(os.PathListSeparator)+env.Getenv("PATH"))

			if fixturePath, ok := os.LookupEnv("FIXTURES"); ok {
				env.Vars = append(env.Vars, "FIXTURES="+fixturePath)
			}

			return nil
		},
	})
}

func getGoEnv() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "env", "-json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run go env: %w", err)
	}

	var envMap map[string]string
	if err := json.Unmarshal(output, &envMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return envMap, nil
}

func getPolicyBinPath() (string, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return "", err
	}

	policyPath := filepath.Join(projectRoot, "dist", policyFileName())

	switch fi, err := os.Stat(policyPath); {
	case err != nil:
		return "", err
	case fi.IsDir():
		return "", errors.Errorf("%q is a directory, not a file", policyPath)
	default:
		return filepath.Abs(policyPath)
	}
}

func getProjectRoot() (string, error) {
	if err := os.Chdir("../.."); err != nil {
		return "", err
	}

	return os.Getwd()
}

const windows = `windows`

func policyFileName() string {
	goEnv, err := getGoEnv()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	b := strings.Builder{}

	b.WriteString("policy" + "_" + runtime.GOOS + "_" + runtime.GOARCH)

	if goARM64, ok := goEnv["GOARM64"]; ok && goARM64 != "" {
		b.WriteString("_" + goARM64)
	}

	if goAMD64, ok := goEnv["GOAMD64"]; ok && goAMD64 != "" {
		b.WriteString("_" + goAMD64)
	}

	b.WriteRune(os.PathSeparator)

	b.WriteString("policy")

	if runtime.GOOS == windows {
		b.WriteString(".exe")
	}

	return b.String()
}
