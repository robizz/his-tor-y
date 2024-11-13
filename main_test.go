package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robizz/his-tor-y/conf"
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

	conf := conf.Config{
		ExitNode: conf.ExitNode{
			DownloadURLTemplate: fakeURLTemplate,
		},
	}

	args := []string{
		"",
		"now",
		"2024-01",
		"2024-01",
	}

	r := mainReturnWithCode(conf, args)
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
	conf := conf.Config{
		ExitNode: conf.ExitNode{
			DownloadURLTemplate: fakeURLTemplate,
		},
	}

	args := []string{
		"",
		"now",
		"2024-01",
		"2024-01",
	}

	r := mainReturnWithCode(conf, args)
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
	conf := conf.Config{
		ExitNode: conf.ExitNode{
			DownloadURLTemplate: fakeURLTemplate,
		},
	}

	args := []string{
		"",
		"now",
		"2024-01",
		"2024-01",
	}

	r := mainReturnWithCode(conf, args)
	if r != 1 {
		t.Errorf("unexpected return code: %d", r)
	}
}
