package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aserto-dev/runtime"
	"github.com/opcr-io/policy/pkg/parser"
	"github.com/open-policy-agent/opa/repl"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Repl(ref string, maxErrors int) error {
	defer c.Cancel()

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()
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

		err = ociStore.LoadIndex()
		if err != nil {
			return err
		}

		existingRefs = ociStore.ListReferences()
		existingRefParsed, err := parser.CalculatePolicyRef(ref, c.Configuration.DefaultDomain)
		if err != nil {
			return err
		}
		descriptor, ok = existingRefs[existingRefParsed]
		if !ok {
			return errors.Errorf("ref [%s] not found in the local store", ref)
		}
	}

	bundleFile := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", "sha256", descriptor.Digest.Hex())

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

	err = opaRuntime.PluginsManager.Start(c.Context)
	if err != nil {
		return errors.Wrap(err, "OPA runtime failed to start")
	}

	err = opaRuntime.WaitForPlugins(c.Context, time.Minute*1)
	if err != nil {
		return errors.Wrap(err, "plugins didn't start on time")
	}

	loop := repl.New(opaRuntime.PluginsManager.Store, c.Configuration.ReplHistoryFile(), c.UI.Output(), "", maxErrors, fmt.Sprintf("running policy [%s]", ref))
	loop.Loop(context.Background())

	return nil
}
