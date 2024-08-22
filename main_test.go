package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestUnmarshall tests that our unmarshal business logic is working correctly.
func TestUnmarshall(t *testing.T) {

	someString := `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:55
ExitNode 23B49521BDC4588C7CCF3C38E552504118326B66
Published 2024-01-30 05:44:30
LastStatus 2024-01-30 11:00:00
ExitAddress 194.26.192.64 2024-01-30 11:30:06`

	r := strings.NewReader(someString)
	b := bufio.NewReader(r)
	exitNodes, err := unmarshall(b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if exitNodes[0].ExitNode != "FE39F07EBE7870DCE124AB30DF3ABD0700A43F75" {
		t.Fatalf("not expected")
	}
	if exitNodes[1].ExitNode != "23B49521BDC4588C7CCF3C38E552504118326B66" {
		t.Fatalf("not expected")
	}
	if exitNodes[0].ExitAddresses[0].ExitAddress != "185.241.208.231" {
		t.Fatalf("not expected")
	}
	u, _ := time.Parse(time.RFC3339, "2024-01-30T10:21:54Z")
	if exitNodes[0].ExitAddresses[0].UpdatedAt != u {
		t.Fatalf("not expected")
	}
	if exitNodes[0].ExitAddresses[1].ExitAddress != "185.241.208.232" {
		t.Fatalf("not expected")
	}
	u, _ = time.Parse(time.RFC3339, "2024-01-30T10:21:55Z")
	if exitNodes[0].ExitAddresses[1].UpdatedAt != u {
		t.Fatalf("not expected")
	}
}

// TestMapToMostRecentEntries tests that we are getting multiple updates for each node in input
// but giving just the most updated one as a result.
func TestMapToMostRecentEntries(t *testing.T) {

	// create 2 reders for 2 files and test the update.
	someString := `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:55
ExitNode 23B49521BDC4588C7CCF3C38E552504118326B66
Published 2024-01-30 05:44:30
LastStatus 2024-01-30 11:00:00
ExitAddress 194.26.192.64 2024-01-30 11:30:06`

	r := strings.NewReader(someString)
	b := bufio.NewReader(r)
	exitNodes, err := unmarshall(b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if exitNodes[0].ExitAddresses[0].ExitAddress != "185.241.208.231" {
		t.Fatalf("expected 185.241.208.231 as it is the last ExitAddress update for ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75")
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
