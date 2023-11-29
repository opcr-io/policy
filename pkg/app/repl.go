package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aserto-dev/runtime"
	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/parser"
	"github.com/opcr-io/policy/pkg/errors"
	"github.com/open-policy-agent/opa/repl"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *PolicyApp) Repl(ref string, maxErrors int) error {
	defer c.Cancel()

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	existingRefs, err := ociClient.ListReferences()
	if err != nil {
		return err
	}
	existingRefParsed, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	descriptor, ok := existingRefs[existingRefParsed]

	if !ok {
		err := c.Pull(ref)
		if err != nil {
			return err
		}

		existingRefs, err := ociClient.ListReferences()
		if err != nil {
			return err
		}
		existingRefParsed, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
		if err != nil {
			return err
		}
		descriptor, ok = existingRefs[existingRefParsed]
		if !ok {
			return errors.NotFound.WithMessage("policy [%s] not in the local store", ref)
		}
	}

	// check for media type - if manifest get tarbarll digest hex.
	bundleHex, err := c.getBundleHex(ociClient, &descriptor)
	if err != nil {
		return err
	}

	bundleFile := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", "sha256", bundleHex)

	opaRuntime, cleanup, err := runtime.NewRuntime(c.Context, c.Logger, &runtime.Config{
		InstanceID: "policy-run",
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{bundleFile},
		},
	})
	if err != nil {
		return errors.ReplFailed.WithError(err)
	}
	defer cleanup()

	err = opaRuntime.Start(c.Context)
	if err != nil {
		return errors.ReplFailed.WithError(err)
	}

	err = opaRuntime.WaitForPlugins(c.Context, time.Minute*1)
	if err != nil {
		return errors.ReplFailed.WithError(err)
	}

	loop := repl.New(opaRuntime.GetPluginsManager().Store, c.Configuration.ReplHistoryFile(), c.UI.Output(), "", maxErrors, fmt.Sprintf("running policy [%s]", ref))
	loop.Loop(context.Background())

	return nil
}

func (c *PolicyApp) getBundleHex(ociClient *oci.Oci, descriptor *ocispec.Descriptor) (string, error) {
	var bundleHex string
	// check for media type - if manifest get tarbarll digest hex.
	if descriptor.MediaType == ocispec.MediaTypeImageManifest {
		bundleDescriptor, _, err := ociClient.GetTarballAndConfigLayerDescriptor(c.Context, descriptor)
		if err != nil {
			return "", err
		}
		bundleHex = bundleDescriptor.Digest.Hex()
		if bundleHex == "" {
			return "", errors.ReplFailed.WithMessage("current manifest does not contain a MediaTypeImageLayerGzip")
		}
	} else {
		bundleHex = descriptor.Digest.Hex()
	}

	return bundleHex, nil
}
