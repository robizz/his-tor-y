package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
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

// TestUnmarshall tests that our unmarshal business logic is working correctly.
func TestUnmarshall(t *testing.T) {

	var someString = `
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

// TestUnmarshall tests that our unmarshal business logic is working correctly.
func TestUnmarshallErrorsIfBadDateFormat(t *testing.T) {

	tests := []struct {
		torNodes              string
		expectedErrorContains string
	}{
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published ERROR-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
`,
			"field Published date parse error:",
		},
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus ERROR-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
`,
			"field LastStatus date parse error:",
		},
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 ERROR-01-30 10:21:54
`,
			"field ExitAddress date parse error:",
		},
	}

	for _, tt := range tests {
		t.Run("should error with: "+tt.expectedErrorContains, func(t *testing.T) {
			r := strings.NewReader(tt.torNodes)
			b := bufio.NewReader(r)
			_, err := unmarshall(b)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("marshall expected error should contain %v", tt.expectedErrorContains)
			}
		})

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

func TestDownloadFile(t *testing.T) {
	// Setup a mock server
	var expected = "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer ts.Close()

	// Setup a temp folder
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}
	defer os.RemoveAll(dir)

	f, err := downloadFile(dir, ts.URL)
	if err != nil {
		t.Errorf("error downloaded file:  %v", err)
	}

	content, err := os.ReadFile(f)
	if err != nil {
		t.Errorf("error reading downloaded file:  %v", err)
	}

	if string(content) != expected {
		t.Errorf("expected %s, but got %s", content, expected)
	}
}

func TestDownloadFileErrorOnDownloadConnection(t *testing.T) {

	// The strategy here is that the mock server answers too late.
	// We set the default timeout for http to be 1ms, but the mock server is going to
	// answer after 10 ms. This makes the downloadFile function to err.
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = 1 * time.Millisecond

	// Setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Setup a temp folder
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}
	// reenableline below once that the code works :)
	defer os.RemoveAll(dir)

	_, err = downloadFile(dir, ts.URL)
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestDownloadFileErrorOnDownload(t *testing.T) {
	// Setup a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// Setup a temp folder
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}
	defer os.RemoveAll(dir)

	_, err = downloadFile(dir, ts.URL)
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestDownloadFileErrorOnMakeTmpFile(t *testing.T) {
	var expected = "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer ts.Close()

	// Setup a temp folder
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setup tmp dir:  %v", err)
	}
	defer os.RemoveAll(dir)

	// Set permission to not allow opening.
	err = os.Chmod(dir, 0000)
	if err != nil {
		t.Errorf("error setup tmp dir permissions:  %v", err)
	}

	_, err = downloadFile(dir, ts.URL)
	if err == nil {
		t.Errorf("error expected")
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

func TestExtractFiles(t *testing.T) {
	// $ tar -tvf test.tar.xz
	// drwxrwxr-x rsora/rsora       0 2024-09-12 18:30 dir1/
	// -rw-rw-r-- rsora/rsora       6 2024-09-12 18:30 dir1/test
	//
	// $ cat dir1/test
	// hello
	var xz = "/Td6WFoAAATm1rRGAgAhARYAAAB0L+Wj4Cf/AIVdADIaSqdFdWDG5DyioorqbKzrYutpz48hW6T+6+aNVA3T8jf0PzyS9ALcmnLhrtM7easSylimqAcho4xEVMQvj0WUss4+rmkoIJai40j22THQcF1sgaTYr2WFsc30TdspFJG2juRj05Obtr1i4YsH5bI9TfNStOkr9x7IyHFMvIuvPA+92QAAAAAA6zfzwvuhqRYAAaEBgFAAAK2nkK2xxGf7AgAAAAAEWVo="
	var expected = "hello"

	dec, err := base64.StdEncoding.DecodeString(xz)
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f, err := os.Create(dir + string(os.PathSeparator) + "test.tar.xz")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	if _, err := f.Write(dec); err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f.Close()

	defer os.RemoveAll(dir)

	err = extractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(dir + string(os.PathSeparator) + "dir1" + string(os.PathSeparator) + "test")
	if err != nil {
		t.Errorf("error reading extracted file:  %v", err)
	}

	// For some weird reason, content contains bunch of spaces and a CR.
	// But I would say this is sufficient to test the happy path :)
	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("expected %s, but got %s.", expected, content)
	}

	// extractFiles should also remove the test.tar.xz file.
	if _, err := os.Stat(dir + string(os.PathSeparator) + "test.tar.xz"); err == nil {
		t.Errorf("tar.xz file should be deleted at this point")
	}
}

func TestExtractFilesErrorOnPermission(t *testing.T) {
	var xz = "/Td6WFoAAATm1rRGAgAhARYAAAB0L+Wj4Cf/AIVdADIaSqdFdWDG5DyioorqbKzrYutpz48hW6T+6+aNVA3T8jf0PzyS9ALcmnLhrtM7easSylimqAcho4xEVMQvj0WUss4+rmkoIJai40j22THQcF1sgaTYr2WFsc30TdspFJG2juRj05Obtr1i4YsH5bI9TfNStOkr9x7IyHFMvIuvPA+92QAAAAAA6zfzwvuhqRYAAaEBgFAAAK2nkK2xxGf7AgAAAAAEWVo="

	dec, err := base64.StdEncoding.DecodeString(xz)
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f, err := os.Create(dir + string(os.PathSeparator) + "test.tar.xz")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	if _, err := f.Write(dec); err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f.Close()

	defer os.RemoveAll(dir)

	// Set permission to not allow opening the tar.xz file.
	err = os.Chmod(dir+string(os.PathSeparator)+"test.tar.xz", 0000)
	if err != nil {
		t.Errorf("error setup tmp file permissions:  %v", err)
	}

	err = extractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
	if err == nil {
		t.Errorf("expected error")
	}

	// Change back permissions and then lock the dir where extraction
	// is supposed to happen.
	err = os.Chmod(dir+string(os.PathSeparator)+"test.tar.xz", 0755)
	if err != nil {
		t.Errorf("error setup tmp file permissions:  %v", err)
	}

	err = os.Chmod(dir, 0000)
	if err != nil {
		t.Errorf("error setup tmp dir permissions:  %v", err)
	}

	err = extractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestExtractFilesErrorMalformedXZ(t *testing.T) {
	var xz = "dGVzdAo="

	dec, err := base64.StdEncoding.DecodeString(xz)
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f, err := os.Create(dir + string(os.PathSeparator) + "test.tar.xz")
	if err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	if _, err := f.Write(dec); err != nil {
		t.Errorf("error setting up tar.xz test: %v", err)
	}

	f.Close()

	defer os.RemoveAll(dir)

	err = extractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
	if err == nil {
		t.Errorf("expected error")
	}
}
