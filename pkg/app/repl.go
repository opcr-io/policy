package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/opcr-io/policy/internal/oci"
	"github.com/opcr-io/policy/internal/parser"
	"github.com/opcr-io/policy/internal/runtime"

	"github.com/opcr-io/policy/pkg/errors"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/repl"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (c *PolicyApp) Repl(ref string, maxErrors int) error {
	defer c.Cancel()

	opaRuntime, err := runtime.New(c.Logger.WithContext(c.Context))
	if err != nil {
		return err
	}

	opaRuntime.Config.InstanceID = "policy-repl"

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	existingRefs, err := ociClient.ListReferences()
	if err != nil {
		return err
	}

	existingRefParsed, err := parser.CalculateRef(ref, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	descriptor, ok := existingRefs[existingRefParsed]

	if !ok {
		err := c.Pull(ref, "")
		if err != nil {
			return err
		}

		existingRefs, err := ociClient.ListReferences()
		if err != nil {
			return err
		}

		existingRefParsed, err := parser.CalculateRef(ref, c.Configuration.DefaultDomain)
		if err != nil {
			return err
		}

		descriptor, ok = existingRefs[existingRefParsed]
		if !ok {
			return errors.ErrNotFound.WithMessage("policy [%s] not in the local store", ref)
		}
	}

	store, err := c.loadStore(ociClient, descriptor)
	if err != nil {
		return err
	}

	outputFormat := ""
	banner := fmt.Sprintf("running policy [%s]", ref)

	r := repl.New(store, ".policy_history", os.Stdout, outputFormat, maxErrors, banner)

	r.Loop(c.Context)

	return nil
}

func (c *PolicyApp) loadStore(ociClient *oci.Oci, descriptor v1.Descriptor) (storage.Store, error) {
	// check for media type - if manifest get tarball digest hex.
	bundleHex, err := c.getBundleHex(ociClient, &descriptor)
	if err != nil {
		return nil, err
	}

	bundleFile := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", "sha256", bundleHex)

	reader, err := os.Open(bundleFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	loader := bundle.NewTarballLoaderWithBaseURL(reader, "")

	bundleReader := bundle.NewCustomReader(loader)

	loadedBundle, err := bundleReader.Read()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := json.Marshal(loadedBundle.Manifest)
	if err != nil {
		return nil, err
	}

	manifest := runtime.MetadataEx{}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, err
	}

	runtime.RegisterStubBuiltins(manifest.Metadata.RequiredBuiltins)

	store := inmem.New()

	txn, err := store.NewTransaction(c.Context, storage.WriteParams)
	if err != nil {
		return nil, err
	}

	opts := bundle.ActivateOpts{
		Ctx:      c.Context,
		Store:    store,
		Txn:      txn,
		Compiler: ast.NewCompiler(),
		Metrics:  metrics.New(),
		Bundles: map[string]*bundle.Bundle{
			"default": &loadedBundle,
		},
	}

	err = bundle.Activate(&opts)
	if err != nil {
		store.Abort(c.Context, txn)
		return nil, err
	}

	if err := store.Commit(c.Context, txn); err != nil {
		return nil, err
	}

	return store, nil
}

func (c *PolicyApp) getBundleHex(ociClient *oci.Oci, descriptor *v1.Descriptor) (string, error) {
	var bundleHex string
	// check for media type - if manifest get tarbarll digest hex.
	if descriptor.MediaType == v1.MediaTypeImageManifest {
		bundleDescriptor, _, err := ociClient.GetTarballAndConfigLayerDescriptor(c.Context, descriptor)
		if err != nil {
			return "", err
		}

		bundleHex = bundleDescriptor.Digest.Hex()
		if bundleHex == "" {
			return "", errors.ErrReplFailed.WithMessage("current manifest does not contain a MediaTypeImageLayerGzip")
		}
	} else {
		bundleHex = descriptor.Digest.Hex()
	}

	return bundleHex, nil
}
