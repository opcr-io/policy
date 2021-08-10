package app

import (
	"io"
	"os"
	"path/filepath"
	"time"

	runtime "github.com/aserto-dev/aserto-runtime"
	"github.com/aserto-dev/aserto-runtime/plugins/edge/builtins"
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

func (c *PolicyApp) Build(refs []string, path string) error {
	defer func() {
		c.Cancel()
	}()

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

	tarball, err := c.buildBundleTgz(workdir, path)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.FileStoreRoot)
	if err != nil {
		return err
	}

	descriptor, err := c.createImage(ociStore, tarball)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		parsed, err := c.calculatePolicyRef(ref)
		if err != nil {
			return err
		}

		ociStore.AddReference(parsed, descriptor)

		c.UI.Normal().
			WithStringValue("reference", ref).
			Msg("Tagging image.")
	}

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) calculatePolicyRef(userRef string) (string, error) {
	parsed, err := docker.ParseDockerRef(userRef)
	if err != nil {
		return "", err
	}

	familiarized := docker.FamiliarString(parsed)

	if familiarized == userRef {
		parsedWithDomain, err := docker.ParseDockerRef(c.Configuration.DefaultDomain + "/" + userRef)
		if err != nil {
			return "", err
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

func (c *PolicyApp) buildBundleTgz(workdir, bundleDir string) (string, error) {
	builtins.Register(c.Logger, c.Runtime.Directory)
	ctx := c.Context

	err := c.Runtime.PluginsManager.Start(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to start OPA plugin manager")
	}

	// TODO: fix this (don't really need plugins?)
	err = c.Runtime.WaitForPlugins(ctx, time.Duration(10)*time.Second)
	if err != nil {
		return "", errors.Wrap(err, "failed to wait for OPA plugins")
	}

	outfile := filepath.Join(workdir, "bundle.tgz")
	err = c.Runtime.Build(runtime.BuildParams{
		Capabilities: &runtime.CapabilitiesFlag{C: nil},
		OutputFile:   outfile,
		BundleMode:   true,
		Target:       util.NewEnumFlag(compile.TargetRego, []string{compile.TargetRego, compile.TargetWasm}),
	}, []string{bundleDir})
	if err != nil {
		return "", errors.Wrap(err, "failed to build policy")
	}

	return outfile, nil
}
