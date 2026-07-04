//go:build ignore

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

type archiveKind int

const (
	archiveTarGz archiveKind = iota
	archiveZip
)

type target struct {
	goos   string
	goarch string
	suffix string
	kind   archiveKind
}

var targets = []target{
	{"darwin", "arm64", "darwin-arm64", archiveTarGz},
	{"linux", "amd64", "linux-amd64", archiveTarGz},
	{"linux", "arm64", "linux-arm64", archiveTarGz},
	{"windows", "amd64", "windows-amd64", archiveZip},
	{"windows", "arm64", "windows-arm64", archiveZip},
}

func main() {
	outDir := "dist"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fatalf("mkdir dist: %v", err)
	}

	var archiveNames []string

	for _, t := range targets {
		binName := "status-line"
		if t.goos == "windows" {
			binName += ".exe"
		}
		binPath := filepath.Join(outDir, binName)

		fmt.Printf("building status-line-%s...\n", t.suffix)

		cmd := exec.Command("go", "build", "-o", binPath, ".")
		cmd.Env = append(os.Environ(),
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fatalf("build status-line-%s: %v", t.suffix, err)
		}

		var archiveName string
		var err error
		switch t.kind {
		case archiveTarGz:
			archiveName = "status-line-" + t.suffix + ".tar.gz"
			err = writeTarGz(filepath.Join(outDir, archiveName), binPath, binName)
		case archiveZip:
			archiveName = "status-line-" + t.suffix + ".zip"
			err = writeZip(filepath.Join(outDir, archiveName), binPath, binName)
		}
		if err != nil {
			fatalf("archive status-line-%s: %v", t.suffix, err)
		}

		if err := os.Remove(binPath); err != nil {
			fatalf("remove intermediate binary %s: %v", binPath, err)
		}

		fmt.Printf("  -> %s\n", filepath.Join(outDir, archiveName))
		archiveNames = append(archiveNames, archiveName)
	}

	if err := writeChecksums(outDir, archiveNames); err != nil {
		fatalf("write checksums: %v", err)
	}
	fmt.Printf("  -> %s\n", filepath.Join(outDir, "SHA256SUMS.txt"))
}

// writeTarGz packages the file at binPath into a gzip-compressed tar archive
// at archivePath, storing it under nameInArchive with executable permissions.
func writeTarGz(archivePath, binPath, nameInArchive string) error {
	data, err := os.ReadFile(binPath)
	if err != nil {
		return err
	}

	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	hdr := &tar.Header{
		Name: nameInArchive,
		Mode: 0755,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(data); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	return f.Close()
}

// writeZip packages the file at binPath into a zip archive at archivePath,
// storing it under nameInArchive with executable permissions.
func writeZip(archivePath, binPath, nameInArchive string) error {
	data, err := os.ReadFile(binPath)
	if err != nil {
		return err
	}

	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	hdr := &zip.FileHeader{
		Name:   nameInArchive,
		Method: zip.Deflate,
	}
	hdr.SetMode(0755)

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}
	return f.Close()
}

// writeChecksums computes the SHA-256 of each named file in dir and writes
// dist/SHA256SUMS.txt in standard `sha256sum` format (hash, two spaces,
// filename only — no directory prefix).
func writeChecksums(dir string, names []string) error {
	sort.Strings(names)

	var out []byte
	for _, name := range names {
		sum, err := sha256File(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		out = fmt.Appendf(out, "%s  %s\n", sum, name)
	}

	return os.WriteFile(filepath.Join(dir, "SHA256SUMS.txt"), out, 0644)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "build: "+format+"\n", args...)
	os.Exit(1)
}
