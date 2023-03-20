package app

import (
	"bytes"
	"io"
	"os"

	"github.com/opcr-io/policy/pkg/oci"
	"github.com/opcr-io/policy/pkg/parser"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func (c *PolicyApp) Save(userRef, outputFilePath string) error {
	defer c.Cancel()
	var outputFile *os.File

	ref, err := parser.CalculatePolicyRef(userRef, c.Configuration.DefaultDomain)
	if err != nil {
		return err
	}

	ociClient, err := oci.NewOCI(c.Context, c.Logger, c.getHosts, c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	refs, err := ociClient.ListReferences()
	if err != nil {
		return err
	}

	refDescriptor, ok := refs[ref]
	if !ok {
		return errors.Errorf("policy [%s] not found in the local store", ref)
	}

	// if the refrence descriptor is the manifest get the tarball descriptor information from the manifest layers
	if refDescriptor.MediaType == ocispec.MediaTypeImageManifest {
		bundleDescriptor, _, err := ociClient.GetTarballAndConfigLayerDescriptor(c.Context, &refDescriptor)
		if err != nil {
			return err
		}
		refDescriptor = *bundleDescriptor
	}

	if outputFilePath == "-" {
		outputFile = os.Stdout
	} else {
		c.UI.Normal().
			WithStringValue("digest", refDescriptor.Digest.String()).
			Msgf("Resolved ref [%s].", ref)
		outputFile, err = os.Create(outputFilePath)

		if err != nil {
			return errors.Wrapf(err, "failed to create output file [%s]", outputFilePath)
		}

		defer func() {
			err := outputFile.Close()
			if err != nil {
				c.UI.Problem().WithErr(err).Msg("Failed to close policy bundle tarball.")
			}
		}()
	}

	err = c.writePolicy(ociClient, &refDescriptor, outputFile)
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) writePolicy(ociStore *oci.Oci, refDescriptor *ocispec.Descriptor, outputFile io.Writer) error {
	reader, err := ociStore.GetStore().Fetch(c.Context, *refDescriptor)
	if err != nil {
		return err
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close OCI policy reader.")
		}
	}()

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(reader)
	if err != nil && err != io.EOF {
		return err
	}
	_, err = outputFile.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
