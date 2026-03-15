package app

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opcr-io/policy/oci"
	perr "github.com/opcr-io/policy/pkg/errors"
	"github.com/opcr-io/policy/pkg/x"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ExtractPolicyBundle extracts a policy bundle from OCI store to the specified directory.
//
//nolint:gocognit,funlen // Security checks require comprehensive validation logic.
func (c *PolicyApp) ExtractPolicyBundle(ociClient *oci.Oci, ref string, destDir string) error {
	// Get reference descriptor
	refDescriptor, err := c.getRefDescriptor(ociClient, ref)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err).WithMessage("failed to get reference descriptor")
	}

	// Fetch the tarball from OCI store
	reader, err := ociClient.GetStore().Fetch(c.Context, *refDescriptor)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err).WithMessage("failed to fetch policy bundle")
	}

	// Wrap the reader depending on media type: gzip-compressed layers use a gzip reader,
	// while plain tar layers are passed through directly.
	var gzReader io.ReadCloser
	if refDescriptor.MediaType == v1.MediaTypeImageLayerGzip {
		gzReader, err = gzip.NewReader(reader)
		if err != nil {
			reader.Close() //nolint:errcheck
			return perr.ErrExtractFailed.WithError(err).WithMessage("failed to create gzip reader")
		}

		defer func() {
			if closeErr := reader.Close(); closeErr != nil {
				c.UI.Problem().WithErr(closeErr).Msg("Failed to close OCI policy reader")
			}
		}()
	} else {
		gzReader = reader
	}

	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			c.UI.Problem().WithErr(closeErr).Msg("Failed to close policy reader")
		}
	}()

	// Get absolute path for security checks
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err).WithMessage("failed to get absolute path")
	}

	// Validate that the destination directory exists
	stat, err := os.Stat(absDestDir)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err).WithMessage("destination directory [%s] does not exist", absDestDir)
	}

	if !stat.IsDir() {
		return perr.ErrExtractFailed.WithMessage("[%s] is not a directory", absDestDir)
	}

	// Extract tar archive
	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return perr.ErrExtractFailed.WithError(err).WithMessage("failed to read tar header")
		}

		// Security check: sanitize and validate header.Name before use
		// This prevents path traversal attacks (CWE-22)
		cleanedName := filepath.Clean(header.Name)

		// Reject absolute paths (check both platform-native and Unix-style for tar archives)
		if filepath.IsAbs(cleanedName) || strings.HasPrefix(header.Name, "/") {
			return perr.ErrExtractFailed.WithMessage("unsafe absolute path in archive: %s", header.Name)
		}

		// Reject paths that escape the destination directory
		if strings.HasPrefix(cleanedName, ".."+string(filepath.Separator)) || cleanedName == ".." {
			return perr.ErrExtractFailed.WithMessage("unsafe path traversal detected: %s", header.Name)
		}

		// Construct target path with sanitized name
		targetPath := filepath.Join(absDestDir, cleanedName)

		// Final safety check: ensure resolved path is within destination
		if !isPathSafe(targetPath, absDestDir) {
			return perr.ErrExtractFailed.WithMessage("unsafe path detected: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, x.OwnerReadWriteExecute); err != nil {
				return perr.ErrExtractFailed.WithError(err).WithMessage("failed to create directory [%s]", targetPath)
			}

		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(targetPath), x.OwnerReadWriteExecute); err != nil {
				return perr.ErrExtractFailed.WithError(err).WithMessage("failed to create parent directory")
			}

			// Create and write file
			outFile, err := os.Create(targetPath)
			if err != nil {
				return perr.ErrExtractFailed.WithError(err).WithMessage("failed to create file [%s]", targetPath)
			}

			// Copy file content
			//nolint:gosec // G110: Controlled tar extraction from trusted OCI registry, not user input.
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return perr.ErrExtractFailed.WithError(err).WithMessage("failed to write file content")
			}

			if err := outFile.Close(); err != nil {
				c.UI.Problem().WithErr(err).Msgf("Failed to close file [%s]", targetPath)
			}

		case tar.TypeSymlink:
			// Policy bundles do not use symlinks. Skip them to avoid symlink-based
			// path traversal attacks (CWE-22, CWE-59).
			c.UI.Problem().Msgf("Skipping symlink [%s] -> [%s]: symlinks are not supported in policy bundles",
				header.Name, header.Linkname)

		default:
			c.UI.Problem().Msgf("Skipping unknown file type %v for [%s]", header.Typeflag, header.Name)
		}
	}

	return nil
}

// isPathSafe checks if absTarget is within absAllowed.
// Both arguments must already be absolute, clean paths (as produced by filepath.Join or filepath.Abs).
// This prevents path traversal attacks.
func isPathSafe(absTarget, absAllowed string) bool {
	rel, err := filepath.Rel(absAllowed, absTarget)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}
