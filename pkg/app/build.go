package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aserto-dev/runtime"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/reference/docker"

	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	orasoci "oras.land/oras-go/v2/content/oci"
)

const (
	AnnotationPolicyRegistryType = "org.openpolicyregistry.type"
	PolicyTypePolicy             = "policy"
)

func (c *PolicyApp) Build(ref string, path []string, annotations map[string]string,
	runConfigFile string,
	target string,
	optimizationLevel int,
	entrypoints []string,
	revision string,
	ignore []string,
	capabilities string,
	verificationKey string,
	verificationKeyID string,
	algorithm string,
	scope string,
	excludeVerifyFiles []string,
	signingKey string,
	claimsFile string) error {
	defer c.Cancel()

	// Create a tmp dir where to do our work.
	workdir, err := os.MkdirTemp("", "policy-build")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary build directory")
	}
	defer func() {
		err := os.RemoveAll(workdir)
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to remove temporary working directory.")
		}
	}()

	opaRuntime, cleanup, err := runtime.NewRuntime(c.Context, c.Logger, &runtime.Config{
		InstanceID: "policy-build",
	})
	if err != nil {
		return errors.Wrap(err, "failed to setup the OPA runtime")
	}
	defer cleanup()

	outfile := filepath.Join(workdir, "bundle.tgz")

	err = opaRuntime.Build(&runtime.BuildParams{
		CapabilitiesJSONFile: capabilities,
		Target:               runtime.Rego,
		OptimizationLevel:    optimizationLevel,
		Entrypoints:          entrypoints,
		OutputFile:           outfile,
		Revision:             revision,
		Ignore:               ignore,
		Debug:                c.Logger.Debug().Enabled(),
		Algorithm:            algorithm,
		Key:                  signingKey,
		Scope:                scope,
		ClaimsFile:           claimsFile,
		PubKey:               verificationKey,
		PubKeyID:             verificationKeyID,
		ExcludeVerifyFiles:   excludeVerifyFiles,
	}, path)
	if err != nil {
		return errors.Wrap(err, "failed to build opa policy bundle")
	}

	ociStore, err := orasoci.New(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	if annotations == nil {
		annotations = map[string]string{}
	}

	familiarezedRef, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	parsedRef, err := docker.ParseDockerRef(familiarezedRef)
	if err != nil {
		return err
	}

	annotations[ocispec.AnnotationTitle] = docker.TrimNamed(parsedRef).String()
	annotations[AnnotationPolicyRegistryType] = PolicyTypePolicy
	annotations[ocispec.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)

	desc, err := c.createImage(ociStore, outfile, annotations)
	if err != nil {
		return err
	}

	err = ociStore.Tag(c.Context, desc, parsedRef.String())
	if err != nil {
		return err
	}

	c.UI.Normal().WithStringValue("reference", parsedRef.String()).Msg("Tagging image.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) createImage(ociStore *orasoci.Store, tarball string, annotations map[string]string) (ocispec.Descriptor, error) {
	descriptor := ocispec.Descriptor{}
	ociStore.AutoSaveIndex = true
	fDigest, err := c.fileDigest(tarball)
	if err != nil {
		return descriptor, err
	}

	tarballFile, err := os.Open(tarball)
	if err != nil {
		return descriptor, err
	}
	fileInfo, err := tarballFile.Stat()
	if err != nil {
		return descriptor, err
	}
	defer func() {
		err := tarballFile.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close bundle tarball.")
		}
	}()

	descriptor.Digest = fDigest
	descriptor.Size = fileInfo.Size()
	descriptor.Annotations = annotations
	descriptor.MediaType = oci.MediaTypeImageLayer

	exists, err := ociStore.Exists(c.Context, descriptor)
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		return descriptor, err
	}

	if exists {
		// Hack to remove the existing digest until ocistore deleter is implemented
		// https://github.com/oras-project/oras-go/issues/454
		digestPath := filepath.Join(strings.Split(descriptor.Digest.String(), ":")...)
		blob := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", digestPath)
		err = os.Remove(blob)
		if err != nil {
			return descriptor, err
		}
	}

	reader := bufio.NewReader(tarballFile)

	err = ociStore.Push(c.Context, descriptor, reader)
	if err != nil {
		return descriptor, err
	}

	configBytes := []byte(fmt.Sprintf("{\"created\":%q}", time.Now().UTC().Format(time.RFC3339)))
	configDesc := content.NewDescriptorFromBytes(oci.MediaTypeConfig, configBytes)

	err = ociStore.Push(c.Context, configDesc, bytes.NewReader(configBytes))
	if err != nil {
		return descriptor, err
	}

	manifestDesc, err := oras.Pack(c.Context, ociStore, oci.MediaTypeConfig, []ocispec.Descriptor{descriptor}, oras.PackOptions{
		PackImageManifest:   true,
		ManifestAnnotations: descriptor.Annotations,
		ConfigDescriptor:    &configDesc,
	})
	if err != nil {
		return manifestDesc, err
	}

	c.UI.Normal().
		WithStringValue("digest", manifestDesc.Digest.String()).
		Msg("Created new image.")

	return manifestDesc, nil
}

func (c *PolicyApp) fileDigest(file string) (digest.Digest, error) {
	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		err := fd.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close bundle tarball when calculating digest.")
		}
	}()

	fDigest, err := digest.FromReader(fd)
	if err != nil {
		return "", err
	}

	return fDigest, nil
}
