//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

func init() {
	// Set go version for docker builds
	os.Setenv("GO_VERSION", "1.19")
	// Enable docker buildkit capabilities
	os.Setenv("DOCKER_BUILDKIT", "1")
}

// Generate generates all code.
func Generate() error {
	return common.GenerateWith([]string{
		filepath.Dir(deps.GoBinPath("mockgen")),
		filepath.Dir(deps.GoBinPath("wire")),
	})
}

// Build builds all binaries in ./cmd.
func Build() error {
	return common.BuildReleaser()
}

// Release releases the project.
func Release(releaseFilePath string) error {
	if releaseFilePath != "" {
		return common.Release("--config", releaseFilePath)
	}
	return common.Release()
}

// BuildAll builds all binaries in ./cmd for
// all configured operating systems and architectures.
func BuildAll() error {
	return common.BuildAllReleaser()
}

// Lint runs linting for the entire project.
func Lint() error {
	return common.Lint()
}

// Test runs all tests and generates a code coverage report.
func Test() error {
	return common.Test()
}

// DockerImage builds the docker image for the project.
func DockerImage() error {
	version, err := common.Version()
	if err != nil {
		return errors.Wrap(err, "failed to calculate version")
	}

	return common.DockerImage(fmt.Sprintf("policy:%s", version))
}

// DockerPush builds the docker image using all tags specified by sver
// and pushes it to the specified registry
func DockerPush(registry, org string) error {
	tags, err := common.DockerTags(registry, fmt.Sprintf("%s/policy", org))
	if err != nil {
		return err
	}

	version, err := common.Version()
	if err != nil {
		return errors.Wrap(err, "failed to calculate version")
	}

	for _, tag := range tags {
		common.UI.Normal().WithStringValue("tag", tag).Msg("pushing tag")
		err = common.DockerPush(fmt.Sprintf("policy:%s", version), fmt.Sprintf("%s/%s/policy:%s", registry, org, tag))
		if err != nil {
			return err
		}
	}

	return nil
}

// Tag creates a new patch version tag and pushes it if it's not dirty.
func Tag() error {
	newTag, err := common.NextVersion("patch")
	if err != nil {
		return err
	}

	if common.IsDirty(newTag) {
		return errors.New("your tree is in a dirty state")
	}

	newTag = "v" + newTag

	err = common.GitTag(newTag)
	if err != nil {
		return err
	}

	err = common.GitPushTag(newTag, "origin")
	if err != nil {
		return err
	}

	common.UI.Normal().WithStringValue("new-tag", newTag).Msg("created and pushed new tag")
	return nil
}

// Deps installs all dependencies required to build the project.
func Deps() {
	deps.GetAllDeps()
}

// All runs all targets in the appropriate order.
// The targets are run in the following order:
// deps, generate, lint, test, build, dockerImage
func All() error {
	mg.SerialDeps(Deps, Generate, Lint, Test, Build, DockerImage)
	return nil
}
