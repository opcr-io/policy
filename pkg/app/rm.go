package app

import (
	"encoding/json"
	"os"
	"path/filepath"

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
	// attach ref name annotation for comparison.
	ref.Annotations = make(map[string]string)
	ref.Annotations[ocispec.AnnotationRefName] = existingRefParsed

	// err = c.removeFromIndex(&ref)
	// if err != nil {
	// 	return err
	// }

	// Reload ociClient with refreshed index to update reference list.
	ociClient, err = oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	if ref.MediaType != oci.MediaTypeImageLayer {
		return c.removeBasedOnManifest(ociClient, &ref)
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
	tarballStillUsed, err := c.tarballReferencedByOtherManifests(ociClient, &ref)
	if err != nil {
		return err
	}
	// only remove the blob if not used by another reference.
	if removeBlob && !tarballStillUsed {
		// Hack to remove the existing digest until ocistore deleter is implemented
		// https://github.com/oras-project/oras-go/issues/454
		err := ociClient.GetStore().Delete(c.Context, ref)
		// err := oci.RemoveBlob(&ref, c.Configuration.PoliciesRoot())
		if err != nil {
			return err
		}
	}

	c.UI.Normal().
		WithStringValue("reference", existingRef).
		Msg("Removed reference.")

	return nil
}

func (c *PolicyApp) removeBasedOnManifest(ociClient *oci.Oci, ref *ocispec.Descriptor) error {
	tarballDesc, configDesc, err := ociClient.GetTarballAndConfigLayerDescriptor(c.Context, ref)
	if err != nil {
		return err
	}

	// Hack to remove the existing digest until ocistore deleter is implemented
	// https://github.com/oras-project/oras-go/issues/454
	err = ociClient.GetStore().Delete(c.Context, *ref)
	//err = oci.RemoveBlob(ref, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	tarballStillUsed, err := c.tarballReferencedByOtherManifests(ociClient, tarballDesc)
	if err != nil {
		return err
	}

	if !tarballStillUsed {
		err = ociClient.GetStore().Delete(c.Context, *tarballDesc)
		//err = oci.RemoveBlob(tarballDesc, c.Configuration.PoliciesRoot())
		if err != nil {
			return err
		}
	}
	err = ociClient.GetStore().Delete(c.Context, *configDesc)
	//err = oci.RemoveBlob(configDesc, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	return nil
}

func (c *PolicyApp) tarballReferencedByOtherManifests(ociClient *oci.Oci, ref *ocispec.Descriptor) (bool, error) {
	type index struct {
		Version   int                  `json:"schemaVersion"`
		Manifests []ocispec.Descriptor `json:"manifests"`
	}

	var localIndex index
	indexPath := filepath.Join(c.Configuration.PoliciesRoot(), "index.json")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(indexBytes, &localIndex)
	if err != nil {
		return false, err
	}
	for i := range localIndex.Manifests {
		elem := localIndex.Manifests[i]
		if elem.MediaType == oci.MediaTypeImageLayer && elem.Digest == ref.Digest {
			return true, nil
		}
		if elem.MediaType != oci.MediaTypeImageLayer {
			manifest, err := ociClient.GetManifest(&elem)
			if err != nil {
				return false, err
			}
			for _, layer := range manifest.Layers {
				if layer.MediaType == ocispec.MediaTypeImageLayerGzip && layer.Digest == ref.Digest {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (c *PolicyApp) removeFromIndex(ref *ocispec.Descriptor) error {

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

	err = os.WriteFile(indexPath, newIndexBytes, 0664) //nolint:gosec // Keep same permissions.
	if err != nil {
		return err
	}
	return nil
}

func removeFromManifests(manifests []ocispec.Descriptor, ref *ocispec.Descriptor) []ocispec.Descriptor {
	newarray := make([]ocispec.Descriptor, len(manifests)-1)
	k := 0
	for i := range manifests {
		if manifests[i].Digest == ref.Digest && manifests[i].Annotations[ocispec.AnnotationRefName] == ref.Annotations[ocispec.AnnotationRefName] {
			continue
		}
		newarray[k] = manifests[i]
		k++
	}
	return newarray
}
