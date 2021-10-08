package app

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Save(userRef, outputFile string) error {
	defer c.Cancel()
	var o *os.File

	ref, err := c.calculatePolicyRef(userRef)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}

	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	refs := ociStore.ListReferences()

	refDescriptor, ok := refs[ref]
	if !ok {
		return errors.Errorf("policy [%s] not found in the local store", ref)
	}

	if outputFile != "-" {
		c.UI.Normal().
			WithStringValue("digest", refDescriptor.Digest.String()).
			Msgf("Resolved ref [%s].", ref)
	}

	reader, err := ociStore.ReaderAt(c.Context, refDescriptor)
	if err != nil {
		return errors.Wrap(err, "failed to open store reader")
	}

	defer func() {
		err := reader.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close OCI policy reader.")
		}
	}()

	if outputFile == "-" {
		o = os.Stdout
	} else {
		o, err = os.Create(outputFile)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to create output file [%s]", outputFile)
	}
	defer func() {
		err := o.Close()
		if err != nil {
			c.UI.Problem().WithErr(err).Msg("Failed to close policy bundle tarball.")
		}
	}()

	chunkSize := 64
	buf := make([]byte, chunkSize)
	for i := 0; i < int(reader.Size()); {
		if chunkSize < int(reader.Size())-i {
			chunkSize = int(reader.Size()) - i
			buf = make([]byte, chunkSize)
		}

		n, err := reader.ReadAt(buf, int64(i))
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to read OCI policy content")
		}

		_, err = o.Write(buf[:n])
		if err != nil {
			return errors.Wrap(err, "failed to write policy bundle tarball to file")
		}

		if err == io.EOF {
			break
		}

		i += chunkSize
	}

	return nil
}
