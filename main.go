package main

import (
	"archive/tar"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// TODO:
// time to start testing files are too big
// multiple tars.xz should be downloaded, we are doing this exercise just with one day now
// when treating multiple days, duplicates management needs to be managed.
// a final cleanup of all text files must be done
// don't forget testing
// clean comments
// variable names are ugly

type ExitNode struct {
	ExitNode      string        `json:"ExitNode"`
	Published     string        `json:"Published"`
	LastStatus    string        `json:"LastStatus"`
	ExitAddresses []ExitAddress `json:"ExitAddresses"`
}

type ExitAddress struct {
	ExitAddress string `json:"ExitAddress"`
	UpdatedAt   string `json:"UpdatedAt"`
}

func main() {
	// create main temporary directory
	dir, err := os.MkdirTemp("", "history-")
	if err != nil {
		log.Fatal(err)
	}
	// reenable line below once that the code works :)
	defer os.RemoveAll(dir)

	// open file for download
	f := downloadFile(dir, "https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz")
	err = extractFiles(f)
	if err != nil {
		panic(err)
	}

	files, err := buildFileList(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
		// Opening a file
		file, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		// Creating a Reader and reading the file line by line.
		reader := bufio.NewReader(file)
		exitNodes, err := unmarshall(reader)
		if err != nil {
			panic(err)
		}
		jsonList, err := json.Marshal(&exitNodes)
		if err != nil {
			panic(err)
		}
		fmt.Print(string(jsonList))
		break
	}

	// marshal one file

	// @type tordnsel 1.0
	// Downloaded 2024-01-30 13:02:00
	// ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
	// Published 2024-01-30 00:10:50
	// LastStatus 2024-01-30 10:00:00
	// ExitAddress 185.241.208.231 2024-01-30 10:21:54
	// ExitAddress 185.241.208.232 2024-01-30 10:21:54
	// ExitNode 23B49521BDC4588C7CCF3C38E552504118326B66
	// Published 2024-01-30 05:44:30
	// LastStatus 2024-01-30 11:00:00
	// ExitAddress 194.26.192.64 2024-01-30 11:30:06
	// [...]

}

func unmarshall(r *bufio.Reader) ([]*ExitNode, error) {
	exitNodes := []*ExitNode{}
	var exitNode *ExitNode
	for {
		// Reading a line, lines are short so we don't worry abou getting truncated/prefixes.
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				if exitNode != nil {
					exitNodes = append(exitNodes, exitNode)
				}
				break
			}
			return nil, err
		}

		// here starts marshaller logic
		split := strings.Split(string(line), " ")
		key := split[0]
		// here I'm removing the key from the line to get a number of values (could be 1, 2, or 3 values depending on the entry).
		value := strings.Replace(string(line), key+" ", "", 1)
		// Here I'm splitting the value part, and I'm sure that at least every line type is going to have at least one value.
		values := strings.Split(value, " ")

		switch key {
		//headers, so we skip
		case "@type":
		case "Downloaded":
			continue
		case "ExitNode":
			// If the current ExitNode is not empty, we append it in the list and we move on with a new one.
			if exitNode != nil {
				exitNodes = append(exitNodes, exitNode)
			}
			// Time sto start filling a new ExitNode struct
			exitNode = new(ExitNode)
			exitNode.ExitNode = values[0]
		case "Published":
			exitNode.Published = strings.Join(values, " ")
		case "LastStatus":
			exitNode.LastStatus = strings.Join(values, " ")
		case "ExitAddress":
			e := ExitAddress{
				ExitAddress: values[0],
				UpdatedAt:   values[1] + " " + values[2],
			}
			exitNode.ExitAddresses = append(exitNode.ExitAddresses, e)
		default:
			fmt.Println(key)
		}
	}
	return exitNodes, nil
}

func buildFileList(dir string) ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {

				fileList = append(fileList, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

// Matt Holt uses a "file approach" meaning you pass path to functions that do the magic
// https://github.com/mholt/archiver/blob/cdc68dd1f170b8dfc1a0d2231b5bb0967ed67006/tarxz.go#L53-L66
func downloadFile(dir, uri string) string {
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

func extractFiles(fileURI string) error {

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
			// if extraction is ok delete xz file
			fileHandle.Close()
			err = os.Remove(fileURI)
			if err != nil {
				return err
			}
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
