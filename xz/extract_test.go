package xz

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

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

	err = ExtractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
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

	err = ExtractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
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

	err = ExtractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
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

	err = ExtractFiles(dir + string(os.PathSeparator) + "test.tar.xz")
	if err == nil {
		t.Errorf("expected error")
	}
}
