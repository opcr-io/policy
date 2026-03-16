package tests_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aserto-dev/runtime"
	ilog "github.com/opcr-io/policy/internal/logger"
	"github.com/opcr-io/policy/pkg/app"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/opcr-io/policy/pkg/clui"
	"github.com/opcr-io/policy/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func NewCmdContext(t testing.TB) *cmd.Globals {
	t.Helper()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}

	t.Logf("HOME: %q", homeDir)
	t.Setenv("HOME", homeDir)

	ctx, cancel := context.WithCancel(t.Context())

	logger := zerolog.New(os.Stderr)

	ui := clui.NewUIWithOutputErrorAndInput(os.Stdout, os.Stderr, os.Stdin)

	cfg := cmd.Globals{
		Debug:     false,
		Config:    filepath.Join(homeDir, ".config", "policy", "config.yaml"),
		Verbosity: 0,
		Insecure:  false,
		Plaintext: false,
		App: &app.PolicyApp{
			Context: ctx,
			Cancel:  cancel,
			Logger:  &logger,
			Configuration: &config.Config{
				FileStoreRoot: filepath.Join(homeDir, ".policy"),
				DefaultDomain: "",
				Logging: ilog.Config{
					Prod:           false,
					LogLevelParsed: zerolog.InfoLevel,
					LogLevel:       "info",
					GrpcLogLevel:   "info",
				},
			},
			UI: ui,
		},
	}

	return &cfg
}

type BuildOption func(*cmd.BuildCmd) error

func NewBuildCmd(t testing.TB, opts ...BuildOption) *cmd.BuildCmd {
	t.Helper()

	cmd := &cmd.BuildCmd{
		Tag:                "",
		Path:               []string{},
		Annotations:        map[string]string{},
		RunConfigFile:      "",
		Target:             "",
		OptimizationLevel:  0,
		Entrypoints:        []string{},
		Revision:           "",
		Ignore:             []string{},
		Capabilities:       "",
		VerificationKey:    "",
		VerificationKeyID:  "",
		Algorithm:          "",
		Scope:              "",
		ExcludeVerifyFiles: []string{},
		SigningKey:         "",
		ClaimsFile:         "",
		RegoVersion:        runtime.RegoUndefined.String(),
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func BuildWithTag(tag string) BuildOption {
	return func(cmd *cmd.BuildCmd) error {
		if tag == "" {
			return errors.Errorf("tag cannot be empty")
		}

		cmd.Tag = tag

		return nil
	}
}

func BuildWithSourcePath(src []string) BuildOption {
	return func(cmd *cmd.BuildCmd) error {
		if len(src) == 0 {
			return errors.Errorf("source path cannot be empty")
		}

		cmd.Path = append(cmd.Path, src...)

		return nil
	}
}

func BuildWithRegoVersion(ver runtime.RegoVersion) BuildOption {
	return func(cmd *cmd.BuildCmd) error {
		if ver == runtime.RegoUndefined {
			return errors.Errorf("rego version is undefined")
		}

		cmd.RegoVersion = ver.String()

		return nil
	}
}

type ImagesOption func(*cmd.ImagesCmd) error

func NewImagesCmd(t testing.TB, opts ...ImagesOption) *cmd.ImagesCmd {
	t.Helper()

	cmd := &cmd.ImagesCmd{
		Server:    "",
		ShowEmpty: false,
		Org:       "",
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

type InspectOption func(*cmd.InspectCmd) error

func NewInspectCmd(t testing.TB, opts ...InspectOption) *cmd.InspectCmd {
	t.Helper()

	cmd := &cmd.InspectCmd{
		Policy: "",
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func InspectWithPolicy(policy string) InspectOption {
	return func(cmd *cmd.InspectCmd) error {
		if policy == "" {
			return errors.Errorf("policy is empty")
		}

		cmd.Policy = policy

		return nil
	}
}

type RmOption func(*cmd.RmCmd) error

func NewRmCmd(t testing.TB, opts ...RmOption) *cmd.RmCmd {
	t.Helper()

	cmd := &cmd.RmCmd{
		Policies: []string{},
		All:      false,
		Force:    false,
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func RmWithPolicies(policies []string) RmOption {
	return func(cmd *cmd.RmCmd) error {
		if len(policies) == 0 {
			return errors.Errorf("policies cannot be empty")
		}

		cmd.Policies = append(cmd.Policies, policies...)

		return nil
	}
}

func RmWithAll(all bool) RmOption {
	return func(cmd *cmd.RmCmd) error {
		cmd.All = all

		return nil
	}
}

func RmWithForce(force bool) RmOption {
	return func(cmd *cmd.RmCmd) error {
		cmd.Force = force

		return nil
	}
}

type SaveOption func(*cmd.SaveCmd) error

func NewSaveCmd(t testing.TB, opts ...SaveOption) *cmd.SaveCmd {
	t.Helper()

	cmd := &cmd.SaveCmd{
		Policy: "",
		File:   "",
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func SaveWithPolicy(policy string) SaveOption {
	return func(cmd *cmd.SaveCmd) error {
		if policy == "" {
			return errors.Errorf("policy is empty")
		}

		cmd.Policy = policy

		return nil
	}
}

func SaveWithFile(file string) SaveOption {
	return func(cmd *cmd.SaveCmd) error {
		if file == "" {
			return errors.Errorf("file is empty")
		}

		cmd.File = file

		return nil
	}
}

type VersionOption func(*cmd.VersionCmd) error

func NewVersionCmd(t testing.TB, opts ...VersionOption) *cmd.VersionCmd {
	t.Helper()

	cmd := &cmd.VersionCmd{}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func LogStep(cmd string) {
	fmt.Fprintf(os.Stderr, "\n=> policy %s\n", cmd)
}
