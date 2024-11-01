package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/robizz/his-tor-y/download"
	"github.com/robizz/his-tor-y/exitnode"
	"github.com/robizz/his-tor-y/files"
	"github.com/robizz/his-tor-y/xz"
)

// TODO:
// probably the generate file list should go in the extract package and be a deeper module...
// also the nodes structure make the package parse dumb as a name mmmm
/*
packages proposal:
- command
- download
- extract
- parse
- transform
- output
*/
// command line options to tune the resolution of the compaction
// when treating multiple days, duplicates management needs to be managed.
// a final cleanup of all text files must be done
// are we sure we want to use pointers for exit nodes? for now we have values, maybe a memory footprint and performance instrumentation with a full year of data would be nice
// When program reaches the desired complexity and tests are in place, apply effective go / practical go / bill kennedy refactoring
// clean comments
// variable names are ugly
// create a cache and allow commands to run in the cache (maybe using a bolt db? an embedded database? an in memory struct?)
// the in memory struct could be also a zipped json or array of zipped items of a struct that you decompress on the fly, perf it would be nice.
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

// We do this wrapping to allow all defer()s to run before actually exiting.
func main() {
	var start string
	var end string

	flag.StringVar(&start, "start", "2024-01", "The start month in a range search")
	flag.StringVar(&end, "end", "2024-03", "The end month in a range search")
	flag.Parse()

	os.Exit(mainReturnWithCode(exitListsURLTemplate, start, end))
}

func mainReturnWithCode(urlTemplate, start, end string) int {
	// create main temporary directory
	dir, err := os.MkdirTemp("", "his-tor-y-")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}
	// reenable line below once that the code works :)
	defer os.RemoveAll(dir)

	dates, err := generateYearDashMonthInterval(start, end)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	// fmt.Println(dates)
	// open files for download
	for _, d := range dates {
		u := fmt.Sprintf(urlTemplate, d)
		// fmt.Println(u)
		// Performances can be improved if download happens in parallel.
		f, err := download.DownloadFile(dir, u)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return 1
		}
		// Performances can be improved if extraction happens in parallel.
		err = xz.ExtractFiles(f)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return 1
		}
	}

	// ------

	// this business logic here needs to be refactored into a package that exposes a "reader" obj that tracks all the
	// opened files and exposes a method to close them so that I can embed them in a defer function

	nodeFiles, err := files.NewReader(dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	defer nodeFiles.Close()

	// ---------

	// mapToMostRecentEntries is going to be just a functionality that answers a question like:
	// Is this IP a tor exit node NOW?
	// doing
	// `go run . [with maybe an -all parameter ] > nodes.json && jq '.[] | select(.ExitAddresses[].ExitAddress == "107.189.31.187")' nodes.json
	// of caourse the date parameters for his-tor-y should be configured properly.
	// Performances can probably be improved if read happens in parallel,
	// However dedup happens leveraging an hashmap, so same entries must be accessed
	// in a safe way with some sort of semaphore. Measure performances before and after.
	v, err := mapToMostRecentEntries(nodeFiles.Readers)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	jsonList, err := json.Marshal(&v)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
func mapToMostRecentEntries(readers []*bufio.Reader) ([]exitnode.ExitNode, error) {
	updated := make(map[string]exitnode.ExitNode)
	for _, reader := range readers {

		exitNodes, err := exitnode.Unmarshal(reader)
		if err != nil {
			return nil, fmt.Errorf("unmarshall error for file reader: %w", err)
		}

		for _, n := range exitNodes {
			updated[n.ExitNode] = n
		}

	}

	v := make([]exitnode.ExitNode, 0, len(updated))
	for _, value := range updated {
		v = append(v, value)
	}
	return v, nil
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
