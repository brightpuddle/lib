// Package archive provides a way to handle the archive customer data.
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Archive handles the archive customer data.
type Archive struct {
	TmpDir string
}

func openTar(path string) (arc Archive, err error) {
	f, err := os.Open(path)
	if err != nil {
		return arc, err
	}
	defer f.Close()

	tmpDir, err := os.MkdirTemp("", "batteries-")
	if err != nil {
		return arc, fmt.Errorf("cannot create temp dir: %v", err)
	}

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return arc, err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return arc, err
		}
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(filepath.Join(tmpDir, header.Name), 0o700)
			continue
		}
		dst, err := os.Create(filepath.Join(tmpDir, header.Name))
		if err != nil {
			return arc, err
		}
		defer dst.Close()
		if _, err := io.Copy(dst, tarReader); err != nil {
			return arc, err
		}
	}
	return Archive{TmpDir: tmpDir}, nil
}

func openZip(path string) (arc Archive, err error) {
	// Open zip reader and input archive
	r, err := zip.OpenReader(path)
	if err != nil {
		return Archive{}, err
	}
	defer r.Close()

	tmpDir, err := os.MkdirTemp("", "batteries-")
	if err != nil {
		return arc, fmt.Errorf("cannot create temp dir: %v", err)
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			os.MkdirAll(filepath.Join(tmpDir, f.Name), 0o700)
			continue
		}
		// Handles __MACOS being treated like a file
		if strings.Contains(f.Name, "/") {
			continue
		}
		src, err := f.Open()
		if err != nil {
			return arc, err
		}
		defer src.Close()
		dst, err := os.Create(filepath.Join(tmpDir, f.Name))
		if err != nil {
			return arc, err
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			return arc, err
		}
	}

	return Archive{TmpDir: tmpDir}, nil
}

// Open unzips the archive to a tmp folder and returns an Archive.
func Open(path string) (arc Archive, err error) {
	switch {
	case strings.HasSuffix(path, ".zip"):
		return openZip(path)
	case strings.HasSuffix(path, ".tar.gz"):
		return openTar(path)
	default:
		return arc, fmt.Errorf("unrecognized file format for %s", path)
	}
}

// Close cleans up tmp files.
func (arc Archive) Close() error {
	return os.RemoveAll(arc.TmpDir)
}
