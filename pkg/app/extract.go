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

const maxExtractFileSize = 256 << 20 // 256 MiB per file.

func (c *PolicyApp) ExtractPolicyBundle(ociClient *oci.Oci, ref string, destDir string) error {
	refDescriptor, err := c.getRefDescriptor(ociClient, ref)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err)
	}

	rc, err := ociClient.GetStore().Fetch(c.Context, *refDescriptor)
	if err != nil {
		return perr.ErrExtractFailed.WithError(err)
	}

	defer rc.Close()

	var tarInput io.Reader

	if refDescriptor.MediaType == v1.MediaTypeImageLayerGzip {
		gzReader, gzErr := gzip.NewReader(rc)
		if gzErr != nil {
			return perr.ErrExtractFailed.WithMessage("failed to create gzip reader").WithError(gzErr)
		}

		defer gzReader.Close()

		tarInput = gzReader
	} else {
		tarInput = rc
	}

	// Validate and resolve the destination directory.
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to resolve destination path [%s]", destDir).WithError(err)
	}

	// Resolve symlinks in the destination path itself to prevent
	// symlink-based path traversal (CWE-59).
	absDestDir, err = filepath.EvalSymlinks(absDestDir)
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to resolve destination path [%s]", destDir).WithError(err)
	}

	stat, err := os.Stat(absDestDir)
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("destination directory [%s] does not exist", destDir).WithError(err)
	}

	if !stat.IsDir() {
		return perr.ErrExtractFailed.WithMessage("[%s] is not a directory", destDir)
	}

	tarReader := tar.NewReader(tarInput)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return perr.ErrExtractFailed.WithMessage("failed to read tar entry").WithError(err)
		}

		cleanedName := filepath.Clean(header.Name)

		// Reject absolute paths (platform-native and Unix-style in tar).
		if filepath.IsAbs(cleanedName) || strings.HasPrefix(header.Name, "/") {
			return perr.ErrExtractFailed.WithMessage("absolute path in archive [%s]", header.Name)
		}

		// Reject parent directory traversal.
		if strings.HasPrefix(cleanedName, ".."+string(filepath.Separator)) || cleanedName == ".." {
			return perr.ErrExtractFailed.WithMessage("path traversal in archive [%s]", header.Name)
		}

		targetPath := filepath.Join(absDestDir, cleanedName)

		if !isPathSafe(targetPath, absDestDir) {
			return perr.ErrExtractFailed.WithMessage("path [%s] escapes destination directory", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, x.OwnerReadWriteExecute); err != nil {
				return perr.ErrExtractFailed.WithMessage("failed to create directory [%s]", targetPath).WithError(err)
			}

		case tar.TypeReg:
			if err := c.extractRegularFile(targetPath, absDestDir, tarReader); err != nil {
				return err
			}

		case tar.TypeSymlink:
			c.UI.Normal().Msgf("Skipping symlink [%s] -> [%s]: symlinks are not supported.", header.Name, header.Linkname)

		default:
			c.UI.Normal().Msgf("Skipping unknown file type [%d] for [%s].", header.Typeflag, header.Name)
		}
	}

	return nil
}

func (c *PolicyApp) extractRegularFile(targetPath, absDestDir string, tarReader io.Reader) error {
	parentDir := filepath.Dir(targetPath)

	if err := os.MkdirAll(parentDir, x.OwnerReadWriteExecute); err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to create parent directory for [%s]", targetPath).WithError(err)
	}

	// Verify the resolved parent directory is still within the destination
	// to guard against pre-existing symlinks in intermediate path components (CWE-59).
	resolvedParent, err := filepath.EvalSymlinks(parentDir)
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to resolve parent directory [%s]", parentDir).WithError(err)
	}

	if !isPathSafe(resolvedParent, absDestDir) {
		return perr.ErrExtractFailed.WithMessage("parent directory [%s] resolves outside destination", parentDir)
	}

	// Guard against writing through pre-existing symlinks (CWE-59).
	if lstat, lErr := os.Lstat(targetPath); lErr == nil && lstat.Mode()&os.ModeSymlink != 0 {
		return perr.ErrExtractFailed.WithMessage("refusing to write through symlink [%s]", targetPath)
	}

	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, x.OwnerReadWrite) //nolint:gosec // G304: targetPath validated above
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to create file [%s]", targetPath).WithError(err)
	}

	written, copyErr := io.Copy(outFile, io.LimitReader(tarReader, maxExtractFileSize)) //nolint:gosec // G110: size bounded by LimitReader
	if copyErr != nil {
		outFile.Close()
		return perr.ErrExtractFailed.WithMessage("failed to write file [%s]", targetPath).WithError(copyErr)
	}

	if written >= maxExtractFileSize {
		outFile.Close()
		return perr.ErrExtractFailed.WithMessage("file [%s] exceeds maximum allowed size (%d bytes)", targetPath, maxExtractFileSize)
	}

	if err := outFile.Close(); err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to close file [%s]", targetPath).WithError(err)
	}

	return nil
}

func isPathSafe(absTarget, absAllowed string) bool {
	rel, err := filepath.Rel(absAllowed, absTarget)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}
