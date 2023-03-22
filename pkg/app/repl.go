package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aserto-dev/runtime"
	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/open-policy-agent/opa/repl"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
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
			return errors.Errorf("ref [%s] not found in the local store", ref)
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
		return errors.Wrap(err, "failed to setup the OPA runtime")
	}
	defer cleanup()

	err = opaRuntime.Start(c.Context)
	if err != nil {
		return errors.Wrap(err, "OPA runtime failed to start")
	}

	err = opaRuntime.WaitForPlugins(c.Context, time.Minute*1)
	if err != nil {
		return errors.Wrap(err, "plugins didn't start on time")
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
			return "", errors.New("current manifest does not contain a MediaTypeImageLayerGzip")
		}
	} else {
		bundleHex = descriptor.Digest.Hex()
	}

	return bundleHex, nil
}
