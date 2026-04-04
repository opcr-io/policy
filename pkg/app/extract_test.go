package app_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/containerd/containerd/v2/core/remotes/docker"
	poci "github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/pkg/app"
	"github.com/opcr-io/policy/pkg/clui"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// tarEntry describes a single entry in a test tar archive.
type tarEntry struct {
	Name     string
	Body     string
	TypeFlag byte
	Linkname string
}

func buildTar(t *testing.T, entries []tarEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.Name,
			Mode:     0o644,
			Typeflag: e.TypeFlag,
			Linkname: e.Linkname,
		}

		switch e.TypeFlag {
		case tar.TypeDir:
			hdr.Mode = 0o755
		case tar.TypeReg, 0:
			hdr.Size = int64(len(e.Body))
			hdr.Typeflag = tar.TypeReg
		}

		require.NoError(t, tw.WriteHeader(hdr))

		if e.TypeFlag == tar.TypeReg || e.TypeFlag == 0 {
			_, err := tw.Write([]byte(e.Body))
			require.NoError(t, err)
		}
	}

	require.NoError(t, tw.Close())

	return buf.Bytes()
}

func gzipBytes(t *testing.T, b []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(b)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	return buf.Bytes()
}

func pushBlob(t *testing.T, ctx context.Context, ociClient *poci.Oci, data []byte, mediaType, ref string) {
	t.Helper()

	desc := content.NewDescriptorFromBytes(mediaType, data)
	require.NoError(t, ociClient.GetStore().Push(ctx, desc, bytes.NewReader(data)))
	require.NoError(t, ociClient.GetStore().Tag(ctx, desc, ref))
}

func newTestApp(t *testing.T) *app.PolicyApp {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	logger := zerolog.Nop()
	ui := clui.NewUI()

	return &app.PolicyApp{
		Context: ctx,
		Cancel:  cancel,
		Logger:  &logger,
		UI:      ui,
	}
}

func noopHosts(_ string) ([]docker.RegistryHost, error) {
	return nil, nil
}

func newTestOCI(t *testing.T) *poci.Oci {
	t.Helper()

	ctx := context.Background()
	logger := zerolog.Nop()
	storeDir := t.TempDir()

	ociClient, err := poci.NewOCI(ctx, &logger, noopHosts, storeDir)
	require.NoError(t, err)

	return ociClient
}

// --- isPathSafe unit tests ---

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		allowed   string
		expected  bool
	}{
		{"same directory", "/a/b", "/a/b", true},
		{"child", "/a/b/c", "/a/b", true},
		{"parent", "/a", "/a/b", false},
		{"sibling", "/a/c", "/a/b", false},
		{"traversal", "/a/b/../../c", "/a/b", false},
		{"exact parent with dot-dot", "/a/b/..", "/a/b", false},
		{"nested child", "/a/b/c/d/e", "/a/b", true},
		{"different root", "/x/y", "/a/b", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := app.IsPathSafe(tc.target, tc.allowed)
			require.Equal(t, tc.expected, result)
		})
	}
}

// --- ExtractPolicyBundle tests ---

const testRef = "test.io/policy:latest"

func TestExtractPolicyBundle_PlainTar(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "dir/", TypeFlag: tar.TypeDir},
		{Name: "dir/file.txt", Body: "hello world"},
		{Name: "root.txt", Body: "root content"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(destDir, "dir", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(content))

	content, err = os.ReadFile(filepath.Join(destDir, "root.txt"))
	require.NoError(t, err)
	require.Equal(t, "root content", string(content))
}

func TestExtractPolicyBundle_GzipTar(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "gzipped.txt", Body: "compressed content"},
	})

	pushBlob(t, a.Context, ociClient, gzipBytes(t, tarData), v1.MediaTypeImageLayerGzip, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(destDir, "gzipped.txt"))
	require.NoError(t, err)
	require.Equal(t, "compressed content", string(content))
}

func TestExtractPolicyBundle_SymlinkSkipped(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "real.txt", Body: "real"},
		{Name: "link.txt", TypeFlag: tar.TypeSymlink, Linkname: "real.txt"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	// Real file should exist.
	_, err = os.Stat(filepath.Join(destDir, "real.txt"))
	require.NoError(t, err)

	// Symlink should NOT have been created.
	_, err = os.Lstat(filepath.Join(destDir, "link.txt"))
	require.True(t, os.IsNotExist(err))
}

func TestExtractPolicyBundle_RefNotFound(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	err := a.ExtractPolicyBundle(ociClient, "nonexistent.io/policy:v1", destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "extract failed")
}

func TestExtractPolicyBundle_DestDirDoesNotExist(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)

	tarData := buildTar(t, []tarEntry{
		{Name: "file.txt", Body: "data"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, filepath.Join(t.TempDir(), "nonexistent"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "extract failed")
}

func TestExtractPolicyBundle_AbsolutePathInArchive(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	// Build a tar archive with an absolute path entry.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	_ = tw.WriteHeader(&tar.Header{
		Name:     "/etc/passwd",
		Mode:     0o644,
		Size:     4,
		Typeflag: tar.TypeReg,
	})
	_, _ = tw.Write([]byte("evil"))
	_ = tw.Close()

	pushBlob(t, a.Context, ociClient, buf.Bytes(), v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "absolute path")
}

func TestExtractPolicyBundle_PathTraversalInArchive(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	// Build a tar archive with a path traversal entry.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	_ = tw.WriteHeader(&tar.Header{
		Name:     "../../outside.txt",
		Mode:     0o644,
		Size:     6,
		Typeflag: tar.TypeReg,
	})
	_, _ = tw.Write([]byte("escape"))
	_ = tw.Close()

	pushBlob(t, a.Context, ociClient, buf.Bytes(), v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path traversal")
}

func TestExtractPolicyBundle_SymlinkEscapeSkipped(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "escape", TypeFlag: tar.TypeSymlink, Linkname: "../../etc/passwd"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	// Symlink should not have been created.
	_, err = os.Lstat(filepath.Join(destDir, "escape"))
	require.True(t, os.IsNotExist(err))
}

func TestExtractPolicyBundle_AbsoluteSymlinkTargetSkipped(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "abslink", TypeFlag: tar.TypeSymlink, Linkname: "/etc/passwd"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	_, err = os.Lstat(filepath.Join(destDir, "abslink"))
	require.True(t, os.IsNotExist(err))
}

func TestExtractPolicyBundle_HardlinkSkipped(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, []tarEntry{
		{Name: "original.txt", Body: "data"},
		{Name: "hardlink.txt", TypeFlag: tar.TypeLink, Linkname: "original.txt"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(destDir, "original.txt"))
	require.NoError(t, err)

	// Hardlink should not have been created.
	_, err = os.Lstat(filepath.Join(destDir, "hardlink.txt"))
	require.True(t, os.IsNotExist(err))
}

func TestExtractPolicyBundle_EmptyTar(t *testing.T) {
	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	tarData := buildTar(t, nil)

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.NoError(t, err)
}

func TestExtractPolicyBundle_PreExistingSymlinkBlocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}

	a := newTestApp(t)
	ociClient := newTestOCI(t)
	destDir := t.TempDir()

	// Create a symlink at the target path pointing outside destDir.
	outsideDir := t.TempDir()
	symPath := filepath.Join(destDir, "trap.txt")
	require.NoError(t, os.Symlink(filepath.Join(outsideDir, "stolen.txt"), symPath))

	// Build a tar that writes to the same path as the symlink.
	tarData := buildTar(t, []tarEntry{
		{Name: "trap.txt", Body: "payload"},
	})

	pushBlob(t, a.Context, ociClient, tarData, v1.MediaTypeImageLayer, testRef)

	err := a.ExtractPolicyBundle(ociClient, testRef, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "refusing to write through symlink")

	// Verify no file was written outside destDir.
	_, err = os.Stat(filepath.Join(outsideDir, "stolen.txt"))
	require.True(t, os.IsNotExist(err))
}
