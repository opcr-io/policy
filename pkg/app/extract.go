package app

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opcr-io/policy/oci"
	perr "github.com/opcr-io/policy/pkg/errors"
	"github.com/opcr-io/policy/pkg/x"
)

// ExtractPolicyBundle extracts a policy bundle from OCI store to the specified directory
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
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			c.UI.Problem().WithErr(closeErr).Msg("Failed to close OCI policy reader")
		}
	}()

	// Create gzip reader
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err).WithMessage("failed to create gzip reader")
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			c.UI.Problem().WithErr(closeErr).Msg("Failed to close gzip reader")
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
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return perr.ErrExtractFailed.WithError(err).WithMessage("failed to read tar header")
		}

		// Security check: prevent path traversal attacks
		targetPath := filepath.Join(absDestDir, header.Name)
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
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return perr.ErrExtractFailed.WithError(err).WithMessage("failed to write file content")
			}

			if err := outFile.Close(); err != nil {
				c.UI.Problem().WithErr(err).Msgf("Failed to close file [%s]", targetPath)
			}

		case tar.TypeSymlink:
			// Handle symlinks carefully - ensure they don't point outside destDir
			linkTarget := header.Linkname

			// Resolve symlink target relative to the file's directory
			symlinkDir := filepath.Dir(targetPath)
			resolvedTarget := filepath.Join(symlinkDir, linkTarget)

			// Security check for symlink target
			if !isPathSafe(resolvedTarget, absDestDir) {
				return perr.ErrExtractFailed.WithMessage("unsafe symlink detected: %s -> %s", header.Name, linkTarget)
			}

			// Create symlink
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				c.UI.Problem().WithErr(err).Msgf("Failed to create symlink [%s]", targetPath)
			}

		default:
			c.UI.Problem().Msgf("Skipping unknown file type %v for [%s]", header.Typeflag, header.Name)
		}
	}

	return nil
}

// isPathSafe checks if the target path is within the allowed directory
// This prevents path traversal attacks
func isPathSafe(targetPath, allowedDir string) bool {
	// Clean and normalize paths
	cleanTarget := filepath.Clean(targetPath)
	cleanAllowed := filepath.Clean(allowedDir)

	// Get absolute paths
	absTarget, err := filepath.Abs(cleanTarget)
	if err != nil {
		return false
	}

	absAllowed, err := filepath.Abs(cleanAllowed)
	if err != nil {
		return false
	}

	// Check if target is within allowed directory
	// Use filepath.Rel to check if target is a subdirectory of allowed
	rel, err := filepath.Rel(absAllowed, absTarget)
	if err != nil {
		return false
	}

	// If rel starts with "..", it's outside the allowed directory
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}
