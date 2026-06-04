package app

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/opcr-io/policy/internal/oci"
	"github.com/opcr-io/policy/internal/parser"
	"github.com/opcr-io/policy/internal/runtime"

	"github.com/containerd/errdefs"
	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orasoci "oras.land/oras-go/v2/content/oci"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	AnnotationPolicyRegistryType = "org.openpolicyregistry.type"
	PolicyTypePolicy             = "policy"
)

//nolint:funlen
func (c *PolicyApp) Build(
	ref string,
	path []string,
	annotations map[string]string,
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
	regoVersion runtime.RegoVersion,
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

	opaRuntime, err := runtime.New(c.Logger.WithContext(c.Context))
	if err != nil {
		return errors.Wrap(err, "failed to setup the OPA runtime")
	}

	opaRuntime.Config.InstanceID = "policy-build"

	outFile := filepath.Join(workDir, "bundle.tgz")

	err = opaRuntime.Build(&runtime.BuildParams{
		CapabilitiesJSONFile: capabilities,
		Target:               runtime.Rego,
		OptimizationLevel:    optimizationLevel,
		Entrypoints:          entrypoints,
		OutputFile:           outFile,
		Revision:             revision,
		Ignore:               ignore,
		Debug:                c.Logger.GetLevel() == zerolog.DebugLevel,
		Algorithm:            algorithm,
		Key:                  signingKey,
		Scope:                scope,
		ClaimsFile:           claimsFile,
		PubKey:               verificationKey,
		PubKeyID:             verificationKeyID,
		ExcludeVerifyFiles:   excludeVerifyFiles,
		RegoVersion:          regoVersion,
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

	parsedRef, err := parser.CalculateNamedRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return errors.Wrap(err, "failed to calculate policy reference")
	}

	annotations = buildAnnotations(annotations, parsedRef, regoVersion)

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

func buildAnnotations(annotations map[string]string, parsedRef reference.Named, regoVersion runtime.RegoVersion) map[string]string {
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[v1.AnnotationTitle] = parsedRef.Name()
	annotations[AnnotationPolicyRegistryType] = PolicyTypePolicy
	annotations[v1.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)
	annotations["rego.version"] = regoVersion.String()

	return annotations
}

func (c *PolicyApp) createImage(ociStore *orasoci.Store, tarball string, annotations map[string]string) (v1.Descriptor, error) {
	ociStore.AutoSaveIndex = true
	ociStore.AutoGC = true

	// tarball layer
	tarDescriptor, err := c.createTarLayer(ociStore, tarball, annotations)
	if err != nil {
		return v1.Descriptor{}, err
	}

	// cfg layer
	cfgDescriptor, err := c.createEmptyCfgLayer(ociStore)
	if err != nil {
		return v1.Descriptor{}, err
	}

	manifestDesc, err := oras.PackManifest(
		c.Context,
		ociStore,
		oras.PackManifestVersion1_1,
		v1.MediaTypeImageManifest,
		oras.PackManifestOptions{
			Layers:              []v1.Descriptor{tarDescriptor},
			ManifestAnnotations: tarDescriptor.Annotations,
			ConfigDescriptor:    &cfgDescriptor,
		},
	)
	if err != nil {
		return v1.Descriptor{}, err
	}

	c.UI.Normal().
		WithStringValue("digest", manifestDesc.Digest.String()).
		Msg("Created new image.")

	return manifestDesc, nil
}

func (c *PolicyApp) createEmptyCfgLayer(ociStore *orasoci.Store) (v1.Descriptor, error) {
	cfg := []byte("{}")

	cfgDescriptor := v1.Descriptor{
		MediaType: v1.MediaTypeEmptyJSON,
		Digest:    digest.FromBytes(cfg),
		Size:      int64(len(cfg)),
	}

	cfgExist, err := ociStore.Exists(c.Context, cfgDescriptor)
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		return v1.Descriptor{}, err
	}

	if err := ociStore.Delete(c.Context, cfgDescriptor); cfgExist && err != nil {
		return v1.Descriptor{}, err
	}

	if err := ociStore.Push(c.Context, cfgDescriptor, bytes.NewReader(cfg)); err != nil {
		return v1.Descriptor{}, err
	}

	cfgDescriptor.Annotations = map[string]string{
		v1.AnnotationCreated: time.Now().UTC().Format(time.RFC3339),
	}

	return cfgDescriptor, nil
}

func (c *PolicyApp) createTarLayer(ociStore *orasoci.Store, tarball string, annotations map[string]string) (v1.Descriptor, error) {
	tarDigest, err := c.fileDigest(tarball)
	if err != nil {
		return v1.Descriptor{}, err
	}

	tarReader, err := os.Open(tarball)
	if err != nil {
		return v1.Descriptor{}, err
	}

	defer func() {
		if err := tarReader.Close(); err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close bundle tarball.")
		}
	}()

	fileInfo, err := tarReader.Stat()
	if err != nil {
		return v1.Descriptor{}, err
	}

	tarDescriptor := v1.Descriptor{
		Digest:      tarDigest,
		Size:        fileInfo.Size(),
		Annotations: annotations,
		MediaType:   oci.MediaTypeImageLayer,
	}

	exists, err := ociStore.Exists(c.Context, tarDescriptor)
	if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
		return v1.Descriptor{}, err
	}

	// delete if exists
	if err := ociStore.Delete(c.Context, tarDescriptor); exists && err != nil {
		return v1.Descriptor{}, err
	}

	reader := bufio.NewReader(tarReader)

	if err := ociStore.Push(c.Context, tarDescriptor, reader); err != nil {
		return v1.Descriptor{}, err
	}

	return tarDescriptor, nil
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
