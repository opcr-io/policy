package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Rm(existingRef string, force bool) error {
	defer c.Cancel()

	existingRefParsed, err := parser.CalculatePolicyRef(existingRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	confirmation := force
	if !force {
		c.UI.Exclamation().
			WithStringValue("reference", existingRefParsed).
			WithAskBoolMap("[Y/n]", &confirmation, map[string]bool{
				"":  true,
				"y": true,
				"n": false,
			}).Msgf("Are you sure?")
	}

	if !confirmation {
		c.UI.Exclamation().Msg("Operation canceled by user.")
		return nil
	}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	existingRefs, err := ociClient.ListReferences()
	if err != nil {
		return err
	}

	ref, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", existingRef)
	}
	// attach ref name annotation for comparisson.
	ref.Annotations = make(map[string]string)
	ref.Annotations[ocispec.AnnotationRefName] = existingRefParsed

	err = c.removeFromIndex(ref)
	if err != nil {
		return err
	}

	// Reload ociClient with refreshed index to update reference list.
	ociClient, err = oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	updatedRefs, err := ociClient.ListReferences()
	if err != nil {
		return err
	}
	// Check if existing images use same digest.
	removeBlob := true
	for _, v := range updatedRefs {
		if v.Digest == ref.Digest {
			removeBlob = false
			break
		}
	}
	// only remove the blob if not used by another reference.
	if removeBlob {
		// Hack to remove the existing digest until ocistore deleter is implemented
		// https://github.com/oras-project/oras-go/issues/454
		digestPath := filepath.Join(strings.Split(ref.Digest.String(), ":")...)
		blob := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", digestPath)
		err = os.Remove(blob)
		if err != nil {
			return err
		}
	}

	c.UI.Normal().
		WithStringValue("reference", existingRef).
		Msg("Removed reference.")

	return nil
}

func (c *PolicyApp) removeFromIndex(ref ocispec.Descriptor) error {

	type index struct {
		Version   int                  `json:"schemaVersion"`
		Manifests []ocispec.Descriptor `json:"manifests"`
	}

	var localIndex index
	indexPath := filepath.Join(c.Configuration.PoliciesRoot(), "index.json")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(indexBytes, &localIndex)
	if err != nil {
		return err
	}

	localIndex.Manifests = removeFromManifests(localIndex.Manifests, ref)

	newIndexBytes, err := json.Marshal(localIndex)
	if err != nil {
		return err
	}

	err = os.WriteFile(indexPath, newIndexBytes, 0664)
	if err != nil {
		return err
	}
	return nil
}

func removeFromManifests(manifests []ocispec.Descriptor, ref ocispec.Descriptor) []ocispec.Descriptor {
	newarray := make([]ocispec.Descriptor, len(manifests)-1)
	k := 0
	for i := 0; i < len(manifests); i++ {
		if manifests[i].Digest == ref.Digest && manifests[i].Annotations[ocispec.AnnotationRefName] == ref.Annotations[ocispec.AnnotationRefName] {
			continue
		}
		newarray[k] = manifests[i]
		k++
	}
	return newarray
}
