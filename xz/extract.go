package xz

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

func ExtractFiles(ctx context.Context, fileURI string) error {
	// here debug and find a place where toiplement the contect cancellation
	fileHandle, err := os.Open(fileURI)
	if err != nil {
		return fmt.Errorf("extract file error: %w", err)
	}
	defer fileHandle.Close()

	r, err := xz.NewReader(fileHandle)
	if err != nil {
		return fmt.Errorf("xz reader error: %w", err)
	}

	// untar
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		switch {
		// no more files
		case err == io.EOF:
			// if extraction is ok delete xz file
			fileHandle.Close()
			err = os.Remove(fileURI)
			if err != nil {
				return fmt.Errorf("tar reader error: %w", err)
			}
			return nil
		case err != nil:
			return fmt.Errorf("tar reader error: %w", err)
		case header == nil:
			continue
		}

		target := filepath.Join(filepath.Dir(fileURI), header.Name)

		switch header.Typeflag {
		// create directory if doesn't exit
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return fmt.Errorf("tar reader error: %w", err)

				}
			}
		// create file
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("tar reader error: %w", err)
			}
			defer f.Close()

			// copy contents to file
			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("tar reader error: %w", err)
			}
		}
	}
}
