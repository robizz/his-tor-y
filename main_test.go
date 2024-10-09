package main

import (
	"bufio"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestMainReturnWithCode is the integration test for the happy path.
func TestMainReturnWithCode(t *testing.T) {

	var happyxz = "/Td6WFoAAATm1rRGBMCcA4AcIQEWAAAAAAAAAJTfEwfgDf8BlF0AMp4JVwAUv4o0Se2uOQoAeCa9bRsjuAxO7ensztcweQ4vqTehTm70VrFwC56JobMMJA9pN0hxEJrISH3UM2Gco3oCpSgxdhJqF4pvwovzXIU3pVsHrxclP+Mwf+18s6Jqit760tO+pq174ynpfWaFG5jpwmeBn2l0owK0B27vhSBWjUzOEq/pJwAtPnTiOXeY0Fh0rpnuo8PRgVnIfktlbeS9jaXfy/QS81SgRNZu8CGQZeW4CQRT3N8Iam+AdW1Ri7XgnHymeRVkH822u1QxDCLWdcnRJVn/oKmQRmo5MVhNUkuNPAwmGO+wdQQ/zL++cQEISzcRzs3gwD4RT8psHR7iOsewrw++o/tBU3IhgB5ZxmSukVJgv3FvaHgSVbBzGd6+91DdB+ZgsQpokMUKOV6rr+1AmhBEKPOXee28CteivwAJ+9xPMWuHYpzAOtNrkBBg6Gjx48Ceqtd+dyT7q5fPgxgvWg3PbF7TI75xSacmsDcccNPSaaL7QskmNQU0Gv+30g7rCdvmkExu4CTZVGbqsgAAeiBdKNN5JMYAAbgDgBwAADrTCYqxxGf7AgAAAAAEWVo="
	dec, err := base64.StdEncoding.DecodeString(happyxz)
	if err != nil {
		t.Errorf("error setup server:  %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(dec)
	}))
	defer ts.Close()

	fakeURLTemplate := ts.URL + "/%s"
	r := mainReturnWithCode(fakeURLTemplate, "2024-01", "2024-01")
	if r != 0 {
		t.Errorf("unexpected return code: %d", r)
	}
}

// TestMainReturnWithCodeErrorOnDownload is the integration test for download error.
func TestMainReturnWithCodeErrorOnDownload(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	fakeURLTemplate := ts.URL + "/%s"
	r := mainReturnWithCode(fakeURLTemplate, "2024-01", "2024-01")
	if r != 1 {
		t.Errorf("unexpected return code: %d", r)
	}
}

// TestMainReturnWithCodeErrorMalformedXZ is the integration test for extraction error.
func TestMainReturnWithCodeErrorMalformedXZ(t *testing.T) {

	var xz = "dGVzdAo="
	dec, err := base64.StdEncoding.DecodeString(xz)
	if err != nil {
		t.Errorf("error setup server:  %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(dec)
	}))
	defer ts.Close()

	fakeURLTemplate := ts.URL + "/%s"
	r := mainReturnWithCode(fakeURLTemplate, "2024-01", "2024-01")
	if r != 1 {
		t.Errorf("unexpected return code: %d", r)
	}
}

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

func TestBuildFileList(t *testing.T) {
	dir1, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}

	dir2, err := os.MkdirTemp(dir1, "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}

	_, err = os.Create(dir2 + string(os.PathSeparator) + "file1")
	if err != nil {
		t.Errorf("error setup tmp dir2:  %v", err)
	}

	_, err = os.Create(dir2 + string(os.PathSeparator) + "file2")
	if err != nil {
		t.Errorf("error setup tmp dir2:  %v", err)
	}

	defer os.RemoveAll(dir1)

	tree, err := buildFileList(dir1)
	if err != nil {
		t.Errorf("error buildFileList:  %v", err)
	}

	//I'm also checking order here
	if tree[0] != dir2+string(os.PathSeparator)+"file1" {
		t.Errorf("expected %s but got %s", dir1+string(os.PathSeparator)+dir2+string(os.PathSeparator)+"file1", tree[0])
	}

	if tree[1] != dir2+string(os.PathSeparator)+"file2" {
		t.Errorf("expected %s but got %s", dir1+string(os.PathSeparator)+dir2+string(os.PathSeparator)+"file2", tree[1])
	}
}

func TestBuildFileListError(t *testing.T) {
	dir1, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}

	defer os.RemoveAll(dir1)

	// Set permission to not allow opening.
	err = os.Chmod(dir1, 0000)
	if err != nil {
		t.Errorf("error setup tmp dir permissions:  %v", err)
	}

	_, err = buildFileList(dir1)
	if err == nil {
		t.Errorf("error expected")
	}

}


