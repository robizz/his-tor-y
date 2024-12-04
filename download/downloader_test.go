package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

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

	f, err := DownloadFile(context.Background(), dir, ts.URL)
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

	_, err = DownloadFile(context.Background(), dir, ts.URL)
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

	_, err = DownloadFile(context.Background(), dir, ts.URL)
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

	_, err = DownloadFile(context.Background(), dir, ts.URL)
	if err == nil {
		t.Errorf("error expected")
	}
}
