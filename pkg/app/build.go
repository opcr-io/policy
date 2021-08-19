package app

import (
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/policy/pkg/opa"
	containerd_content "github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/reference/docker"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/compile"
	"github.com/open-policy-agent/opa/util"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Build(ref string, path []string,
	runConfigFile string,
	target string,
	optimizationLevel int,
	entrypoints []string,
	revision string,
	bundleMode bool,
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

	params := opa.BuildParams{
		BundleMode:         bundleMode,
		OptimizationLevel:  optimizationLevel,
		Revision:           revision,
		Ignore:             ignore,
		Debug:              c.Logger.Debug().Enabled(),
		Algorithm:          algorithm,
		ExcludeVerifyFiles: excludeVerifyFiles,
		PubKey:             verificationKey,
		PubKeyID:           verificationKeyID,
		Key:                signingKey,
		Scope:              scope,
		ClaimsFile:         claimsFile,
	}

	params.Target = util.NewEnumFlag(compile.TargetRego, []string{compile.TargetRego, compile.TargetWasm})
	err = params.Target.Set(target)
	if err != nil {
		return errors.Wrap(err, "invalid value for target flag")
	}

	params.Capabilities = &opa.CapabilitiesFlag{}
	params.Capabilities.Set(capabilities)
	if err != nil {
		return errors.Wrap(err, "invalid value for capabilities flag")
	}

	params.Entrypoints = opa.RepeatedStringFlag{}
	for _, e := range entrypoints {
		params.Entrypoints.Set(e)
	}

	tarball, err := c.buildBundleTgz(workdir, params, path)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	descriptor, err := c.createImage(ociStore, tarball)
	if err != nil {
		return err
	}

	parsed, err := c.calculatePolicyRef(ref)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	ociStore.AddReference(parsed, descriptor)

	c.UI.Normal().
		WithStringValue("reference", ref).
		Msg("Tagging image.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) calculatePolicyRef(userRef string) (string, error) {
	parsed, err := docker.ParseDockerRef(userRef)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse reference [%s]", userRef)
	}

	familiarized := docker.FamiliarString(parsed)

	domain := docker.Domain(parsed)

	if (familiarized == userRef || familiarized == userRef+":latest") && domain == DefaultCanonicalDomain {
		parsedWithDomain, err := docker.ParseDockerRef(c.Configuration.DefaultDomain + "/" + userRef)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse normalized reference [%s]", c.Configuration.DefaultDomain+"/"+userRef)
		}

		return parsedWithDomain.String(), nil
	}

	return userRef, nil
}

func (c *PolicyApp) createImage(ociStore *content.OCIStore, tarball string) (ocispec.Descriptor, error) {
	descriptor := ocispec.Descriptor{}

	fDigest, err := c.fileDigest(tarball)
	if err != nil {
		return descriptor, err
	}

	existingInfo, err := ociStore.Info(c.Context, fDigest)
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		return descriptor, err
	}

	if err == nil {
		descriptor.Digest = existingInfo.Digest
		descriptor.Size = existingInfo.Size

		c.UI.Normal().
			WithStringValue("digest", descriptor.Digest.String()).
			Msg("Using existing image.")
	} else {
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
			Annotations: map[string]string{},
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
	}

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

func (c *PolicyApp) buildBundleTgz(workdir string, params opa.BuildParams, bundleDirs []string) (string, error) {
	outfile := filepath.Join(workdir, "bundle.tgz")

	params.OutputFile = outfile

	err := opa.Build(params, bundleDirs)
	if err != nil {
		return "", errors.Wrap(err, "failed to build policy")
	}

	return outfile, nil
}
