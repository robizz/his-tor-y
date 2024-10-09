package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Matt Holt uses a "file approach" meaning you pass path to functions that do the magic
// https://github.com/mholt/archiver/blob/cdc68dd1f170b8dfc1a0d2231b5bb0967ed67006/tarxz.go#L53-L66
func DownloadFile(dir, uri string) (string, error) {

	fileURI := filepath.Join(dir, path.Base(uri))
	// fmt.Println(fileURI)

	fileHandle, err := os.OpenFile(fileURI, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	defer fileHandle.Close()

	resp, err := http.Get(uri)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download error, server returned %d", resp.StatusCode)
	}

	_, err = io.Copy(fileHandle, resp.Body)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}

	return fileURI, nil
}
