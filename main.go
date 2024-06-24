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
	"time"

	"github.com/ulikunitz/xz"
)

// TODO:
// We need to start some refactorng to extract compaction function from main and allow for unit testing.
// in the future I would give command line options to tune the resolution of the compaction
// multiple tars.xz should be downloaded, we are doing this exercise just with one day now
// when treating multiple days, duplicates management needs to be managed.
// a final cleanup of all text files must be done
// are we sure we want to use pointers for exit nodes? for now we have values, maybe a memory footprint and performance instrumentation with a full year of data would be nice
// When program reaches the desired complexity and tests are in place, apply effective go / practical go / bill kennedy refactoring
// don't forget testing
// clean comments
// variable names are ugly

type ExitNode struct {
	ExitNode      string        `json:"ExitNode"`
	Published     time.Time     `json:"Published"`
	LastStatus    time.Time     `json:"LastStatus"`
	ExitAddresses []ExitAddress `json:"ExitAddresses"`
}

type ExitAddress struct {
	ExitAddress string    `json:"ExitAddress"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

func main() {
	// create main temporary directory
	dir, err := os.MkdirTemp("", "history-")
	if err != nil {
		log.Fatal(err)
	}
	// reenable line below once that the code works :)
	// defer os.RemoveAll(dir)

	// open file for download
	f := downloadFile(dir, "https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz")
	err = extractFiles(f)
	if err != nil {
		panic(err)
	}

	filenames, err := buildFileList(dir)
	if err != nil {
		panic(err)
	}
	// fmt.Print(files)

	// fmt.Println(file)
	// Opening a file
	// Creating a Reader and reading the file line by line.
	// this basically acts as remove duplicates
	//print as list
	// We created a dict, key is ExitNode AKA node id, and we leverage files being ordered by date during
	// filepath.walk so that looping through all files should generate a map containing nodes with last update
	// for each node.
	readers := make([]*bufio.Reader, len(filenames))
	for _, filename := range filenames {

		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		reader := bufio.NewReader(file)
		readers =append(readers, reader)
		// test if this defer actually works.
		defer file.Close()
	}
	
	v := mapToMostRecentEntries(readers)

	jsonList, err := json.Marshal(&v)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(jsonList))

}

func mapToMostRecentEntries(readers []*bufio.Reader) []ExitNode {
	updated := make(map[string]ExitNode)
	for _, reader := range readers {
		
		exitNodes, err := unmarshall(reader)
		if err != nil {
			panic(err)
		}

		for _, n := range exitNodes {
			updated[n.ExitNode] = n
		}

	}

	v := make([]ExitNode, 0, len(updated))
	for _, value := range updated {
		v = append(v, value)
	}
	return v
}

func unmarshall(r *bufio.Reader) ([]ExitNode, error) {
	exitNodes := []ExitNode{}
	var exitNode ExitNode
	for {
		// Reading a line, lines are short so we don't worry about getting truncated/prefixes.
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				if exitNode.ExitNode != "" {
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
			if exitNode.ExitNode != "" {
				exitNodes = append(exitNodes, exitNode)
			}
			// Time sto start filling a new ExitNode struct
			exitNode = ExitNode{}
			exitNode.ExitNode = values[0]
		case "Published":
			u, err := time.Parse(time.RFC3339, values[0]+"T"+values[1]+"Z")
			if err != nil {
				return nil, err
			}
			exitNode.Published = u
		case "LastStatus":
			u, err := time.Parse(time.RFC3339, values[0]+"T"+values[1]+"Z")
			if err != nil {
				return nil, err
			}
			exitNode.LastStatus = u
		case "ExitAddress":
			u, err := time.Parse(time.RFC3339, values[1]+"T"+values[2]+"Z")
			if err != nil {
				return nil, err
			}
			e := ExitAddress{
				ExitAddress: values[0],
				UpdatedAt:   u,
			}
			exitNode.ExitAddresses = append(exitNode.ExitAddresses, e)
		default:
			// skip
			// fmt.Println(key)
		}
	}
	return exitNodes, nil
}

// buildFileList recursively walks inside a folder to generate the list of all
// files inside a folder tree. Items in the list comes out ordered.
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
	// fmt.Println(fileURI)
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
