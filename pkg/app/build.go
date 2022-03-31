package app

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aserto-dev/runtime"
	containerd_content "github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/google/uuid"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
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

	// Create a tmp dir where to do our work
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

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[ocispec.AnnotationTitle] = ref
	annotations[extendedregistry.AnnotationPolicyRegistry] = "policy"
	annotations[ocispec.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)

	descriptor, err := c.createImage(ociStore, outfile, annotations)
	if err != nil {
		return err
	}

	parsed, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	ociStore.AddReference(parsed, descriptor)

	c.UI.Normal().WithStringValue("reference", ref).Msg("Tagging image.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) createImage(ociStore *content.OCIStore, tarball string, annotations map[string]string) (ocispec.Descriptor, error) {
	descriptor := ocispec.Descriptor{}

	fDigest, err := c.fileDigest(tarball)
	if err != nil {
		return descriptor, err
	}

	_, err = ociStore.Info(c.Context, fDigest)
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		return descriptor, err
	}

	if err == nil {
		err = ociStore.Delete(c.Context, fDigest)
		if err != nil {
			return descriptor, errors.Wrap(err, "couldn't overwrite existing image")
		}
	}

	tarballFile, err := os.Open(tarball)
	if err != nil {
		return descriptor, err
	}
	defer func() {
		err := tarballFile.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close bundle tarball.")
		}
	}()

	fileInfo, err := tarballFile.Stat()
	if err != nil {
		return descriptor, err
	}

	descriptor = ocispec.Descriptor{
		MediaType:   MediaTypeImageLayer,
		Digest:      fDigest,
		Size:        fileInfo.Size(),
		Annotations: annotations,
	}

	ociWriter, err := ociStore.Writer(
		c.Context,
		containerd_content.WithDescriptor(descriptor),
		containerd_content.WithRef(uuid.NewString()))
	if err != nil {
		return descriptor, err
	}
	defer func() {
		err := ociWriter.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close local OCI store.")
		}
	}()

	_, err = io.Copy(ociWriter, tarballFile)
	if err != nil {
		return descriptor, err
	}

	err = ociWriter.Commit(c.Context, fileInfo.Size(), fDigest)
	if err != nil {
		return descriptor, err
	}

	c.UI.Normal().
		WithStringValue("digest", ociWriter.Digest().String()).
		Msg("Created new image.")

	return descriptor, nil
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
