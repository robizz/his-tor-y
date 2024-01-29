package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

// TODO:
// look at io.writer https://www.youtube.com/watch?v=A1MS2LHcPuE&ab_channel=GolangCafe and take a note
// figure out how this package works https://github.com/ulikunitz/xz see mholt https://github.com/mholt/archiver/blob/cdc68dd1f170b8dfc1a0d2231b5bb0967ed67006/tarxz.go#L53-L66
// after decompressing we get a tar that should be untarred https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
// a final cleanup of all text files must be done
// don't forget testing

func main() {
	f:= downloadFile("https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz")
	extractFiles(f)
}

func downloadFile(uri string) string{

	dir, err := os.MkdirTemp("", "history-")
	if err != nil {
		log.Fatal(err)
	}
	//defer os.RemoveAll(dir)

	fileURI := filepath.Join(dir, path.Base(uri))
	fmt.Println(fileURI)
	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fileHandle, err := os.OpenFile(fileURI, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer fileHandle.Close()

	_, err = io.Copy(fileHandle, resp.Body)
	if err != nil {
		panic(err)
	}

	return fileURI
}

func extractFiles(fileURI string) error{


	fileHandle, err := os.Open(fileURI)
	if err != nil {
		panic(err)
	}
	defer fileHandle.Close()
	
	r, err := xz.NewReader(fileHandle)
	if err != nil {
		log.Fatalf("xz Reader error %s", err)
	}
	
	// untar
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		switch {
		// no more files
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(filepath.Dir(fileURI), header.Name)

		switch header.Typeflag {
		// create directory if doesn't exit
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// create file
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy contents to file
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}




}
