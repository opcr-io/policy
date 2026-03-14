package app_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/containerd/containerd/v2/core/remotes/docker"
	"github.com/opcr-io/policy/oci"
	"github.com/opcr-io/policy/pkg/app"
	"github.com/opcr-io/policy/pkg/clui"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content"
)

type tarEntry struct {
	name     string
	content  string
	typeFlag byte
	linkname string // for TypeSymlink
}

func buildTar(entries []tarEntry) []byte {
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)

	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeFlag,
			Linkname: e.linkname,
			Mode:     0o600,
			Size:     int64(len(e.content)),
		}
		if e.typeFlag == tar.TypeDir {
			hdr.Mode = 0o700
			hdr.Size = 0
		}

		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}

		if e.content != "" {
			if _, err := io.WriteString(tw, e.content); err != nil {
				panic(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(b); err != nil {
		panic(err)
	}

	if err := gz.Close(); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func pushBlob(t *testing.T, ctx context.Context, ociClient *oci.Oci, data []byte, mediaType, ref string) {
	t.Helper()

	desc := content.NewDescriptorFromBytes(mediaType, data)

	if err := ociClient.GetStore().Push(ctx, desc, bytes.NewReader(data)); err != nil {
		t.Fatalf("push blob: %v", err)
	}

	if err := ociClient.GetStore().Tag(ctx, desc, ref); err != nil {
		t.Fatalf("tag blob: %v", err)
	}
}

func newTestApp(t *testing.T) (*app.PolicyApp, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	logger := zerolog.Nop()
	ui := clui.NewUIWithOutputErrorAndInput(io.Discard, io.Discard, strings.NewReader(""))

	app := &app.PolicyApp{
		Context: ctx,
		Cancel:  cancel,
		Logger:  &logger,
		UI:      ui,
	}

	return app, cancel
}

func newTestOCI(t *testing.T, ctx context.Context) (*oci.Oci, string) {
	t.Helper()

	storeDir := t.TempDir()
	logger := zerolog.Nop()

	noopHosts := func(string) ([]docker.RegistryHost, error) {
		return nil, nil
	}

	ociClient, err := oci.NewOCI(ctx, &logger, noopHosts, storeDir)
	require.NoError(t, err)

	return ociClient, storeDir
}

func TestIsPathSafe(t *testing.T) {
	base := "/safe/base"

	tests := []struct {
		name     string
		target   string
		allowed  string
		wantSafe bool
	}{
		{
			name:     "same directory",
			target:   "/safe/base",
			allowed:  "/safe/base",
			wantSafe: true,
		},
		{
			name:     "file inside allowed dir",
			target:   "/safe/base/file.txt",
			allowed:  base,
			wantSafe: true,
		},
		{
			name:     "nested directory inside allowed dir",
			target:   "/safe/base/sub/dir/file.rego",
			allowed:  base,
			wantSafe: true,
		},
		{
			name:     "sibling directory",
			target:   "/safe/other",
			allowed:  base,
			wantSafe: false,
		},
		{
			name:     "parent directory",
			target:   "/safe",
			allowed:  base,
			wantSafe: false,
		},
		{
			name:     "root directory",
			target:   "/",
			allowed:  base,
			wantSafe: false,
		},
		{
			name:     "allowed dir is prefix but not parent",
			target:   "/safe/base-extra/file.txt",
			allowed:  base,
			wantSafe: false,
		},
		{
			name:     "dot-dot traversal resolved",
			target:   "/safe/base/sub/../../other",
			allowed:  base,
			wantSafe: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := app.IsPathSafe(filepath.Clean(tc.target), filepath.Clean(tc.allowed))
			require.Equal(t, tc.wantSafe, got)
		})
	}
}

func TestExtractPolicyBundle_PlainTar(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	entries := []tarEntry{
		{name: "subdir/", typeFlag: tar.TypeDir},
		{name: "subdir/policy.rego", typeFlag: tar.TypeReg, content: "package test"},
		{name: "data.json", typeFlag: tar.TypeReg, content: `{"key":"value"}`},
	}

	tarData := buildTar(entries)
	ref := "test/policy:plain"
	pushBlob(t, app.Context, ociClient, tarData, ocispec.MediaTypeImageLayer, ref)

	require.NoError(t, app.ExtractPolicyBundle(ociClient, ref, destDir))

	require.FileExists(t, filepath.Join(destDir, "subdir", "policy.rego"))
	require.FileExists(t, filepath.Join(destDir, "data.json"))

	got, err := os.ReadFile(filepath.Join(destDir, "subdir", "policy.rego"))
	require.NoError(t, err)
	require.Equal(t, "package test", string(got))
}

func TestExtractPolicyBundle_GzipTar(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	entries := []tarEntry{
		{name: "bundle.rego", typeFlag: tar.TypeReg, content: "package bundle"},
	}

	tarData := gzipBytes(buildTar(entries))
	ref := "test/policy:gzip"
	pushBlob(t, app.Context, ociClient, tarData, ocispec.MediaTypeImageLayerGzip, ref)

	require.NoError(t, app.ExtractPolicyBundle(ociClient, ref, destDir))

	got, err := os.ReadFile(filepath.Join(destDir, "bundle.rego"))
	require.NoError(t, err)
	require.Equal(t, "package bundle", string(got))
}

func TestExtractPolicyBundle_Symlink(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	entries := []tarEntry{
		{name: "real.rego", typeFlag: tar.TypeReg, content: "package real"},
		{name: "link.rego", typeFlag: tar.TypeSymlink, linkname: "real.rego"},
	}

	tarData := buildTar(entries)
	ref := "test/policy:symlink"
	pushBlob(t, app.Context, ociClient, tarData, ocispec.MediaTypeImageLayer, ref)

	require.NoError(t, app.ExtractPolicyBundle(ociClient, ref, destDir))

	require.FileExists(t, filepath.Join(destDir, "real.rego"))

	target, err := os.Readlink(filepath.Join(destDir, "link.rego"))
	require.NoError(t, err)
	require.Equal(t, "real.rego", target)
}

func TestExtractPolicyBundle_RefNotFound(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	err := app.ExtractPolicyBundle(ociClient, "nonexistent/ref:latest", destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "extract failed")
}

func TestExtractPolicyBundle_DestDirDoesNotExist(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, _ := newTestOCI(t, app.Context)

	entries := []tarEntry{
		{name: "policy.rego", typeFlag: tar.TypeReg, content: "package p"},
	}
	tarData := buildTar(entries)
	ref := "test/policy:nodest"
	pushBlob(t, app.Context, ociClient, tarData, ocispec.MediaTypeImageLayer, ref)

	err := app.ExtractPolicyBundle(ociClient, ref, "/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
	require.Contains(t, err.Error(), "extract failed")
}

func TestExtractPolicyBundle_AbsolutePathInArchive(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	// Manually craft a tar with an absolute path (bypassing filepath.Clean)
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "/etc/passwd",
		Typeflag: tar.TypeReg,
		Mode:     0o600,
		Size:     4,
	}))
	_, err := io.WriteString(tw, "evil")
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	ref := "test/policy:abspath"
	pushBlob(t, app.Context, ociClient, buf.Bytes(), ocispec.MediaTypeImageLayer, ref)

	err = app.ExtractPolicyBundle(ociClient, ref, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsafe")
}

func TestExtractPolicyBundle_PathTraversalInArchive(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	// Manually craft a tar with a path traversal entry
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "../../outside.txt",
		Typeflag: tar.TypeReg,
		Mode:     0o600,
		Size:     4,
	}))
	_, err := io.WriteString(tw, "evil")
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	ref := "test/policy:traversal"
	pushBlob(t, app.Context, ociClient, buf.Bytes(), ocispec.MediaTypeImageLayer, ref)

	err = app.ExtractPolicyBundle(ociClient, ref, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsafe")
}

func TestExtractPolicyBundle_SymlinkEscape(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	// Symlink that would point outside destDir
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "escape.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "../../etc/passwd",
	}))
	require.NoError(t, tw.Close())

	ref := "test/policy:symlinkescape"
	pushBlob(t, app.Context, ociClient, buf.Bytes(), ocispec.MediaTypeImageLayer, ref)

	err := app.ExtractPolicyBundle(ociClient, ref, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsafe")
}

func TestExtractPolicyBundle_AbsoluteSymlinkTarget(t *testing.T) {
	app, cancel := newTestApp(t)
	defer cancel()

	ociClient, destDir := newTestOCI(t, app.Context)

	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "badlink",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	}))
	require.NoError(t, tw.Close())

	ref := "test/policy:abssymlink"
	pushBlob(t, app.Context, ociClient, buf.Bytes(), ocispec.MediaTypeImageLayer, ref)

	err := app.ExtractPolicyBundle(ociClient, ref, destDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsafe")
}
