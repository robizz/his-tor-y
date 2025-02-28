package xz

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

func Extract(ctx context.Context, fileURI string) error {
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
		select {
		case <-ctx.Done():
			return errors.New("extraction cancelled")
		default:
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

			// create directory if doesn't exit
			// create file
			// copy contents to file
			err = extractFileOrFolder(fileURI, header, tr)
			if err != nil {
				return fmt.Errorf("file or folder extraction error: %w", err)
			}

		}
	}
}

func extractFileOrFolder(fileURI string, header *tar.Header, tr *tar.Reader) error {
	target := filepath.Join(filepath.Dir(fileURI), header.Name)

	switch header.Typeflag {

	case tar.TypeDir:
		if _, err := os.Stat(target); err != nil {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("tar reader error: %w", err)

			}
		}

	case tar.TypeReg:
		f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("tar reader error: %w", err)
		}
		defer f.Close()

		if _, err := io.Copy(f, tr); err != nil {
			return fmt.Errorf("tar reader error: %w", err)
		}
	}
	return nil
}
