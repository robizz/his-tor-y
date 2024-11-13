package business

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

// TestMapToMostRecentEntries tests that we are getting multiple updates for each node in input
// but giving just the most updated one as a result.
func TestMapToMostRecentEntries(t *testing.T) {
	var readers []*bufio.Reader
	// create 2 reders for 2 files and test the update.
	var first = `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:55`

	var second = `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-31 00:10:50
LastStatus 2024-01-31 10:00:00
ExitAddress 185.241.208.231 2024-01-31 10:21:54
ExitAddress 185.241.208.232 2024-01-31 10:21:55`

	r1 := strings.NewReader(first)
	r2 := strings.NewReader(second)
	readers = append(readers, bufio.NewReader(r1), bufio.NewReader(r2))
	nodes, err := mapToMostRecentEntries(readers)
	if err != nil {
		t.Errorf("unexpected mapToMostRecentEntries error")
	}
	if nodes[0].Published.Day() != 31 {
		t.Errorf("expected 31, got: %d", nodes[0].Published.Day())
	}
}

func TestMapToMostRecentEntriesErrorOnUnmarshall(t *testing.T) {
	var readers []*bufio.Reader
	// create 2 reders for 2 files and test the update.
	var first = `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:55`

	var second = `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published NOTADATE LOLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL
LastStatus 2024-01-31 10:00:00
ExitAddress 185.241.208.231 2024-01-31 10:21:54
ExitAddress 185.241.208.232 2024-01-31 10:21:55`

	r1 := strings.NewReader(first)
	r2 := strings.NewReader(second)
	readers = append(readers, bufio.NewReader(r1), bufio.NewReader(r2))
	_, err := mapToMostRecentEntries(readers)
	if err == nil || !strings.Contains(err.Error(), "unmarshall error for file reader") {
		t.Errorf("error expected")
	}
}

func TestGenerateExitListsURLs(t *testing.T) {
	tests := []struct {
		start, end string
		expected   []string
	}{
		{"2024-01", "2024-03", []string{"2024-01", "2024-02", "2024-03"}},
		{"2024-01", "2024-01", []string{"2024-01"}},
		{"2023-12", "2024-02", []string{"2023-12", "2024-01", "2024-02"}},
		{"2022-12", "2024-01", []string{"2022-12", "2023-01", "2023-02", "2023-03", "2023-04", "2023-05", "2023-06", "2023-07", "2023-08", "2023-09", "2023-10", "2023-11", "2023-12", "2024-01"}}, // Span over three years
		{"2024-02", "2024-02", []string{"2024-02"}},
	}

	for _, tt := range tests {
		t.Run(tt.start+" to "+tt.end, func(t *testing.T) {
			actual, err := generateYearDashMonthInterval(tt.start, tt.end)

			if err != nil && !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("generateExitListsURLs(%s, %s) = %v; expected %v", tt.start, tt.end, actual, tt.expected)
			}
		})
	}

	errTests := []struct {
		start, end           string
		expectedErrorMessage string
	}{
		{"2024-03", "2024-01", "start date is after end date"},
		{"202403", "2024-01", "cannot parse"},
		{"2024 03", "2024-01", "cannot parse"},
		{"2024-03", "2024 01", "cannot parse"},
		{"yadda", "2024-01", "cannot parse"},
		{"", "yadda", "cannot parse"},
		{"2024-03", "yadda", "cannot parse"},
	}

	for _, tt := range errTests {
		t.Run(tt.start+" to "+tt.end, func(t *testing.T) {
			nilResult, err := generateYearDashMonthInterval(tt.start, tt.end)
			// Here the strings.Contains() check is a very weak check, this should be implemented using
			// constant errors: https://dave.cheney.net/2016/04/07/constant-errors.
			if nilResult != nil || !strings.Contains(err.Error(), tt.expectedErrorMessage) {
				t.Errorf("generateExitListsURLs(%s, %s) = %v; error expected %v", tt.start, tt.end, err.Error(), tt.expectedErrorMessage)
			}
		})
	}

}
