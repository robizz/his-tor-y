package read

import (
	"os"
	"testing"
)

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

	d := DirReader{
		Dir: dir1,
	}

	tree, err := d.buildFileList()
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

	d := DirReader{
		Dir: dir1,
	}

	_, err = d.buildFileList()
	if err == nil {
		t.Errorf("error expected")
	}

}
