package main

import (
	"archive/tar"
	"bufio"
	"encoding/json"
	"errors"
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
// test file handling functions and see the coverage increasing
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
// create a cache and allow commands to run in the cache (maybe using a bolt db? an embedded database? an in memory struct?)
// command should be silent to use pipe or output redirect. errors should be on stderr
// errors should be constant errors like dave cheney suggests
// we need an integration test to test the whole flow
// Main functionality is: I give you the list of nodes that were found for the time range with the last update inside the time range.
// another funtionality is "IP History":I give you an IP and a parameter like "days", the tool gives me 0 with formatted list of nodes and dates.
// generate go doc
// END TODO

// exitListsURL contains the template for the exit node compressed files URL.
// The string is supposed to be:
// https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz
const exitListsURLTemplate = "https://collector.torproject.org/archive/exit-lists/exit-list-%s.tar.xz"

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

// We do this wrapping to allow all defer()s to run before actually exiting.
func main() { os.Exit(mainReturnWithCode()) }

func mainReturnWithCode() int {
	// create main temporary directory
	dir, err := os.MkdirTemp("", "his-tor-y-")
	if err != nil {
		fmt.Println("Error: %v", err)
		return 1
	}
	// reenable line below once that the code works :)
	defer os.RemoveAll(dir)

	// start and end emulates TUI params for now.
	start := "2024-01"
	end := "2024-03"

	dates, err := generateYearDashMonthInterval(start, end)
	if err != nil {
		fmt.Println("Error: %v", err)
		return 1
	}

	// fmt.Println(dates)
	// open files for download
	for _, d := range dates {
		u := fmt.Sprintf(exitListsURLTemplate, d)
		// fmt.Println(u)
		// Performances can be improved if download happens in parallel.
		f, err := downloadFile(dir, u)
		if err != nil {
			fmt.Println("Error: %v", err)
			return 1
		}
		// Performances can be improved if extraction happens in parallel.
		err = extractFiles(f)
		if err != nil {
			fmt.Println("Error: %v", err)
			return 1
		}
	}

	filenames, err := buildFileList(dir)
	if err != nil {
		fmt.Println("Error: %v", err)
		return 1
	}
	// fmt.Println(filenames)

	// fmt.Println(file)
	// Opening a file
	// Creating a Reader and reading the file line by line.
	// this basically acts as remove duplicates
	//print as list
	// We created a dict, key is ExitNode AKA node id, and we leverage files being ordered by date during
	// filepath.walk so that looping through all files should generate a map containing nodes with last update
	// for each node.
	files := make([]*os.File, len(filenames))
	readers := make([]*bufio.Reader, len(filenames))
	for i, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("Error: %v", err)
			return 1
		}
		// This should be safe because we use make to create a slice with len(filenames) capacity.
		// append does not work for some reason here. Study why.
		readers[i] = bufio.NewReader(file)
		// We are going to use this list of files to close them all.
		files[i] = file
		// test if this defer actually works.
		// defer file.Close()
	}
	// Close all opened files before exiting
	// is this correct? Investigate.
	for _, file := range files {
		defer file.Close()
	}

	// mapToMostRecentEntries is going to be just a functionality that answers a question like:
	// Is this IP a tor exit node NOW?
	// doing
	// `go run . [with maybe an -all parameter ] > nodes.json && jq '.[] | select(.ExitAddresses[].ExitAddress == "107.189.31.187")' nodes.json
	// of caourse the date parameters for his-tor-y should be configured properly.
	// Performances can probably be improved if read happens in parallel,
	// However dedup happens leveraging an hashmap, so same entries must be accessed
	// in a safe way with some sort of semaphore. Measure performances before and after.
	v, err := mapToMostRecentEntries(readers)
	if err != nil {
		fmt.Println("Error: %v", err)
		return 1
	}

	jsonList, err := json.Marshal(&v)
	if err != nil {
		fmt.Println("Error: %v", err)
		return 1
	}

	// Final print do not comment.
	fmt.Print(string(jsonList))
	return 0

}

// mapToMostRecentEntries read all the files, unmarshals them into a list of entries,
// iterate through the entries putting them in a map using the node as a key.
// This generates a map with the most updated entry for each node leveraging 2 side effects:
// files and entries inside files are ordered from older to newer (thanks to buildFileList() )
func mapToMostRecentEntries(readers []*bufio.Reader) ([]ExitNode, error) {
	updated := make(map[string]ExitNode)
	for _, reader := range readers {

		exitNodes, err := unmarshall(reader)
		if err != nil {
			return nil, fmt.Errorf("unmarshall error for file reader: %w", err)
		}

		for _, n := range exitNodes {
			updated[n.ExitNode] = n
		}

	}

	v := make([]ExitNode, 0, len(updated))
	for _, value := range updated {
		v = append(v, value)
	}
	return v, nil
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
				return nil, fmt.Errorf("field Published date parse error: %w", err)
			}
			exitNode.Published = u
		case "LastStatus":
			u, err := time.Parse(time.RFC3339, values[0]+"T"+values[1]+"Z")
			if err != nil {
				return nil, fmt.Errorf("field LastStatus date parse error: %w", err)
			}
			exitNode.LastStatus = u
		case "ExitAddress":
			u, err := time.Parse(time.RFC3339, values[1]+"T"+values[2]+"Z")
			if err != nil {
				return nil, fmt.Errorf("field ExitAddress date parse error: %w", err)
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

func generateYearDashMonthInterval(start, end string) ([]string, error) {

	// Define the date format.
	const yearDashMonth = "2006-01"

	startDate, err := time.Parse(yearDashMonth, start)
	if err != nil {
		return nil, fmt.Errorf("start date parse error: %w", err)
	}

	endDate, err := time.Parse(yearDashMonth, end)
	if err != nil {
		return nil, fmt.Errorf("end date parse error: %w", err)
	}

	if startDate.After(endDate) {
		// This should be implemented using
		// constant errors: https://dave.cheney.net/2016/04/07/constant-errors.
		return nil, errors.New("start date is after end date")
	}

	var dates []string
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 1, 0) {
		dates = append(dates, d.Format(yearDashMonth))
	}

	return dates, nil

}

// Matt Holt uses a "file approach" meaning you pass path to functions that do the magic
// https://github.com/mholt/archiver/blob/cdc68dd1f170b8dfc1a0d2231b5bb0967ed67006/tarxz.go#L53-L66
func downloadFile(dir, uri string) (string, error) {

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

func extractFiles(fileURI string) error {

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
