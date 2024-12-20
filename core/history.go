package core

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/robizz/his-tor-y/download"
	"github.com/robizz/his-tor-y/exitnode"
	"github.com/robizz/his-tor-y/files"
	"github.com/robizz/his-tor-y/xz"
	"golang.org/x/sync/errgroup"
)

func History(ctx context.Context, DownloadURLTemplate, StartDate, EndDate, IP string) (string, error) {
	// create main temporary directory
	dir, err := os.MkdirTemp("", "his-tor-y-")
	if err != nil {
		return "", err
	}
	// reenable line below once that the code works :)
	defer os.RemoveAll(dir)

	dates, err := generateYearDashMonthInterval(StartDate, EndDate)
	if err != nil {
		return "", err
	}

	var g errgroup.Group

	// fmt.Println(dates)
	// open files for download
	for _, d := range dates {
		d := d // new var per iteration
		g.Go(func() error {
			return pull(ctx, DownloadURLTemplate, d, dir)
		})
	}

	if err := g.Wait(); err != nil {
		return "", err
	}

	nodeFiles, err := files.NewReader(dir)
	if err != nil {
		return "", err

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
	v, err := mapToMostRecentEntries(nodeFiles.Readers, IP)
	if err != nil {
		return "", err

	}

	jsonList, err := json.Marshal(&v)
	if err != nil {
		return "", err
	}

	// Final print do not comment.
	return string(jsonList), nil
}

func pull(ctx context.Context, DownloadURLTemplate string, date string, dir string) error {
	u := fmt.Sprintf(DownloadURLTemplate, date)

	fmt.Println(u)

	f, err := download.DownloadFile(ctx, dir, u)
	if err != nil {
		return err

	}
	fmt.Println(f)
	err = xz.Extract(ctx, f)
	fmt.Printf("done: %s\n", f)
	if err != nil {
		return err
	}
	return nil
}

// mapToMostRecentEntries read all the files, unmarshals them into a list of entries,
// iterate through the entries putting them in a map using the node as a key.
// This generates a map with the most updated entry for each node leveraging 2 side effects:
// files and entries inside files are ordered from older to newer (thanks to buildFileList() )
func mapToMostRecentEntries(readers []*bufio.Reader, IP string) ([]exitnode.ExitNode, error) {
	updated := []exitnode.ExitNode{}
	for _, reader := range readers {

		exitNodes, err := exitnode.Unmarshal(reader)
		if err != nil {
			return nil, fmt.Errorf("unmarshall error for file reader: %w", err)
		}
		for _, n := range exitNodes {
			for _, a := range n.ExitAddresses {
				if a.ExitAddress == IP {
					updated = append(updated, n)
					break
				}
			}
		}

	}

	return updated, nil
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
