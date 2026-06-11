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

const MaxExtractFileSize = 256 << 20 // 256 MiB per file.

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
			if err := ensureSafeDir(targetPath, absDestDir); err != nil {
				return err
			}

		case tar.TypeReg, tar.TypeRegA:
			if err := extractRegularFile(targetPath, absDestDir, tarReader); err != nil {
				return err
			}

		case tar.TypeSymlink:
			c.UI.Normal().Msgf("Skipping symlink [%s] -> [%s]: symlinks are not supported.", header.Name, header.Linkname)

		case tar.TypeLink:
			c.UI.Normal().Msgf("Skipping hardlink [%s] -> [%s]: hardlinks are not supported.", header.Name, header.Linkname)

		default:
			c.UI.Normal().Msgf("Skipping unsupported file type [%q] for [%s].", header.Typeflag, header.Name)
		}
	}

	return nil
}

// ensureSafeDir creates a directory at targetDir, walking each path component
// from absDestDir and rejecting any component that is a symlink. This prevents
// MkdirAll from following pre-existing symlinks to create directories outside
// the destination (CWE-59).
func ensureSafeDir(targetDir, absDestDir string) error {
	rel, err := filepath.Rel(absDestDir, targetDir)
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to resolve relative path for [%s]", targetDir).WithError(err)
	}

	if rel == "." {
		return nil
	}

	current := absDestDir

	for _, component := range strings.Split(rel, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}

		current = filepath.Join(current, component)

		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return perr.ErrExtractFailed.WithMessage("path component [%s] is a symlink", current)
			}

			if !info.IsDir() {
				return perr.ErrExtractFailed.WithMessage("path component [%s] is not a directory", current)
			}

			continue
		}

		if !os.IsNotExist(statErr) {
			return perr.ErrExtractFailed.WithMessage("failed to inspect path component [%s]", current).WithError(statErr)
		}

		if mkErr := os.Mkdir(current, x.OwnerReadWriteExecute); mkErr != nil {
			return perr.ErrExtractFailed.WithMessage("failed to create directory [%s]", current).WithError(mkErr)
		}
	}

	return nil
}

// extractRegularFile writes tar content to targetPath using a temp file + atomic
// rename to avoid TOCTOU symlink races (CWE-59). The parent directory must already
// exist and be verified safe by ensureSafeDir.
func extractRegularFile(targetPath, absDestDir string, tarReader io.Reader) error {
	parentDir := filepath.Dir(targetPath)

	if err := ensureSafeDir(parentDir, absDestDir); err != nil {
		return err
	}

	outFile, err := os.CreateTemp(parentDir, ".extract-*")
	if err != nil {
		return perr.ErrExtractFailed.WithMessage("failed to create temporary file for [%s]", targetPath).WithError(err)
	}

	tempPath := outFile.Name()

	written, copyErr := io.Copy(outFile, io.LimitReader(tarReader, MaxExtractFileSize+1)) //nolint:gosec // G110: size bounded by LimitReader
	if copyErr != nil {
		outFile.Close()
		_ = os.Remove(tempPath)

		return perr.ErrExtractFailed.WithMessage("failed to write file [%s]", targetPath).WithError(copyErr)
	}

	if written > MaxExtractFileSize {
		outFile.Close()
		_ = os.Remove(tempPath)

		return perr.ErrExtractFailed.WithMessage("file [%s] exceeds maximum allowed size (%d bytes)", targetPath, MaxExtractFileSize)
	}

	if err := outFile.Close(); err != nil {
		_ = os.Remove(tempPath)

		return perr.ErrExtractFailed.WithMessage("failed to close file [%s]", targetPath).WithError(err)
	}

	// Reject pre-existing symlinks at the target path before installing.
	if lstat, lErr := os.Lstat(targetPath); lErr == nil && lstat.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(tempPath)

		return perr.ErrExtractFailed.WithMessage("refusing to write through symlink [%s]", targetPath)
	}

	// Atomic rename installs the file without following symlinks.
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath)

		return perr.ErrExtractFailed.WithMessage("failed to install file [%s]", targetPath).WithError(err)
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
