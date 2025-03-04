package command

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/robizz/his-tor-y/conf"
)

func TestParse(t *testing.T) {
	n := NewHistory()
	err := n.Parse(conf.Config{}, []string{"test", "now", "2024-01", "2024-02"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
}

func TestParseErrorOnParsing(t *testing.T) {
	n := NewHistory()
	err := n.Parse(conf.Config{}, []string{"test", "now", "2024-01", "true", "ttt"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
}

func TestExecuteNowCoreCall(t *testing.T) {
	var happyxz = "/Td6WFoAAATm1rRGBMC+AoAYIQEWAAAAAAAAAObJVkbgC/8BNl0AMp4JVwAUv4o0Se2uOQoAeCa9bRsjuAxO7ensztcweQ4vqTehTm70VrFwC56JobMMJA9pN0hxEJrISH3UM2Gco3oCpSgxdhJqF4pvwovzXIU3pVsHrxclP+Mwf+18s6Jqit760tO+pq174ynpfWaFG5jpx22bLjgraUd2kQonthTYPlmYGIQygxzX30Jkv2u5/+8hC3d2+JZvh05FofIrrMEVzwY7ygIAKCp5mGCXWYlhudHpfs95Ijtz+zNg4NIqT6Up/lInYzTxguDrVU0KzwM+qhx/gvaJRdIL1Z5MlAF99NqLEBfKUGVjZVmLHhrKjzhiR+atxa5akoqzW6gvDtFup3sNk2UrY86eKzw4qU5oNg/zu20bPjPjYJUkAc/vpn+1pVLxOH9w/SD36JVqjPpA7QRZAAAAAPbVQkNRMK+bAAHaAoAYAAAH/Fz1scRn+wIAAAAABFla"
	dec, err := base64.StdEncoding.DecodeString(happyxz)
	if err != nil {
		t.Errorf("error setup server:  %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(dec)
	}))
	defer ts.Close()

	fakeURLTemplate := ts.URL + "/%s"
	c := conf.Config{
		ExitNode: conf.ExitNode{
			DownloadURLTemplate: fakeURLTemplate,
		},
	}

	// Test default text output
	gold := `ExitNode	Published	LastStatus	ExitAddress	UpdatedAt
FE39F07EBE7870DCE124AB30DF3ABD0700A43F75	2023-12-31 11:29:15 +0000 UTC	2023-12-31 23:00:00 +0000 UTC	185.241.208.232	2023-12-31 23:17:34 +0000 UTC
FE39F07EBE7870DCE124AB30DF3ABD0700A43F75	2023-12-31 11:29:15 +0000 UTC	2023-12-31 23:00:00 +0000 UTC	171.25.193.25	2023-12-31 23:05:55 +0000 UTC
`
	n := NewHistory()
	err = n.Parse(c, []string{"test", "history", "-start", "2024-01", "-end", "2024-01", "-ip", "185.241.208.232"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
	var buf bytes.Buffer
	err = n.Execute(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Expected nil, got: %v", err)
	}
	if buf.String() != gold {
		t.Fatalf("Expected \n%s, got: \n%s", gold, buf.String())
	}

	// Test text output
	gold = `ExitNode	Published	LastStatus	ExitAddress	UpdatedAt
FE39F07EBE7870DCE124AB30DF3ABD0700A43F75	2023-12-31 11:29:15 +0000 UTC	2023-12-31 23:00:00 +0000 UTC	185.241.208.232	2023-12-31 23:17:34 +0000 UTC
FE39F07EBE7870DCE124AB30DF3ABD0700A43F75	2023-12-31 11:29:15 +0000 UTC	2023-12-31 23:00:00 +0000 UTC	171.25.193.25	2023-12-31 23:05:55 +0000 UTC
`
	n = NewHistory()
	err = n.Parse(c, []string{"test", "history", "-start", "2024-01", "-end", "2024-01", "-ip", "185.241.208.232", "-output", "text"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
	buf.Reset()
	err = n.Execute(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Expected nil, got: %v", err)
	}
	if buf.String() != gold {
		t.Fatalf("Expected %s, got: %s", gold, buf.String())
	}

	// Test json output
	gold = `[{"ExitNode":"FE39F07EBE7870DCE124AB30DF3ABD0700A43F75","Published":"2023-12-31T11:29:15Z","LastStatus":"2023-12-31T23:00:00Z","ExitAddresses":[{"ExitAddress":"185.241.208.232","UpdatedAt":"2023-12-31T23:17:34Z"},{"ExitAddress":"171.25.193.25","UpdatedAt":"2023-12-31T23:05:55Z"}]}]`
	n = NewHistory()
	err = n.Parse(c, []string{"test", "history", "-start", "2024-01", "-end", "2024-01", "-ip", "185.241.208.232", "-output", "json"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
	buf.Reset()
	err = n.Execute(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Expected nil, got: %v", err)
	}
	if buf.String() != gold {
		t.Fatalf("Expected %s, got: %s", gold, buf.String())
	}
}

func TestExecuteErrorOnNowCoreCall(t *testing.T) {
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
	c := conf.Config{
		ExitNode: conf.ExitNode{
			DownloadURLTemplate: fakeURLTemplate,
		},
	}

	n := NewHistory()
	err = n.Parse(c, []string{"test", "now", "2024-01", "2024-01"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
	err = n.Execute(context.Background(), os.Stdout)
	if err == nil {
		t.Fatalf("Expected error, got: nil")
	}
}
