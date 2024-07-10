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
	"github.com/distribution/reference"

	oras "github.com/opcr-io/oras-go/v2"
	"github.com/opcr-io/oras-go/v2/content"
	orasoci "github.com/opcr-io/oras-go/v2/content/oci"
	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/parser"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
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
	claimsFile string,
	regoV1 bool,
) error {
	defer c.Cancel()

	workDir, err := os.MkdirTemp("", "policy-build")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary build directory")
	}
	defer func() {
		err := os.RemoveAll(workDir)
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

	outFile := filepath.Join(workDir, "bundle.tgz")

	err = opaRuntime.Build(&runtime.BuildParams{
		CapabilitiesJSONFile: capabilities,
		Target:               runtime.Rego,
		OptimizationLevel:    optimizationLevel,
		Entrypoints:          entrypoints,
		OutputFile:           outFile,
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
		RegoV1:               regoV1,
	}, path)
	if err != nil {
		return errors.Wrap(err, "failed to build opa policy bundle")
	}

	ociStore, err := orasoci.New(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	if ref == "" {
		ref = "default"
	}

	familiarizedRef, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	parsedRef, err := reference.ParseDockerRef(familiarizedRef)
	if err != nil {
		return err
	}

	annotations = buildAnnotations(annotations, parsedRef, regoV1)

	desc, err := c.createImage(ociStore, outFile, annotations)
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

func buildAnnotations(annotations map[string]string, parsedRef reference.Named, regoV1 bool) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[ocispec.AnnotationTitle] = parsedRef.Name()
	annotations[AnnotationPolicyRegistryType] = PolicyTypePolicy
	annotations[ocispec.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)
	if regoV1 {
		annotations["rego.version"] = "rego.V1"
	}

	return annotations
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

	manifestDesc, err := oras.Pack(c.Context, ociStore, ocispec.MediaTypeImageManifest, []ocispec.Descriptor{descriptor}, oras.PackOptions{PackImageManifest: true, ConfigDescriptor: &configDesc, ManifestAnnotations: descriptor.Annotations})
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
