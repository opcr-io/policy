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

var scripts = []string{
	"tests/cli/001-version.txtar",
	"tests/cli/002-fixtures.txtar",
	"tests/cli/003-build-policy_v1.txtar",
	"tests/cli/100-templates-list.txtar",
	"tests/cli/101-template-apply.txtar",
}

func TestScripts(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Logf("cannot determine current working directory %q", err.Error())
		t.FailNow()
	}

	policyBinPath, err := getPolicyBinPath()
	if err != nil {
		t.Logf("cannot determine policy bin location %q", err.Error())
		t.FailNow()
	}

	_, policyBin := filepath.Split(policyBinPath)

	fixturesDir, err := filepath.Abs(filepath.Join(cwd, "../../tests/fixtures"))
	if err != nil {
		t.Logf("cannot determine absolute path to test fixtures directory %q", err.Error())
		t.FailNow()
	}

	testscript.Run(t, testscript.Params{
		Dir:             "",
		Files:           scripts,
		ContinueOnError: true,
		Setup: func(env *testscript.Env) error {
			binDir := filepath.Join(env.WorkDir, "bin")
			if err := os.MkdirAll(binDir, 0o750); err != nil {
				return err
			}

			if err := os.Symlink(policyBinPath, filepath.Join(binDir, policyBin)); err != nil {
				return err
			}

			t.Logf("symlink %q => %q", policyBinPath, filepath.Join(binDir, policyBin))

			env.Vars = append(env.Vars, "PATH="+binDir+string(os.PathListSeparator)+env.Getenv("PATH"))
			t.Logf("PATH="+binDir+string(os.PathListSeparator), binDir)

			if fixturesEnv, ok := os.LookupEnv("FIXTURES"); ok {
				t.Logf("FIXTURES=%q", fixturesEnv)
				env.Vars = append(env.Vars, "FIXTURES="+fixturesEnv)
			} else {
				env.Vars = append(env.Vars, "FIXTURES="+fixturesDir)
			}

			policyRoot := filepath.Join(env.WorkDir, "policy")
			if err := os.MkdirAll(policyRoot, 0o750); err != nil {
				return err
			}

			env.Vars = append(env.Vars, "POLICY_FILE_STORE_ROOT="+policyRoot)

			policyConfig := filepath.Join(policyRoot, "policy.json")

			r, err := os.Create(policyConfig)
			if err != nil {
				t.Logf("failed to create policy/config.json %q", err.Error())
				t.FailNow()
			}

			_ = r.Close()

			return nil
		},
	})

	t.Cleanup(func() {
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
