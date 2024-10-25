package read

import (
	"bufio"
	"os"
	"path/filepath"
)

// this package shouldbe ok at least from the concept perspective.
// you need to fix the tests to use this new concept
// then you need to figure out a good naming for stuff here.

// DirReader is a struct that helps scenning a dir and giving back byte readers
// while keeping reference of al the opened files and readers
type DirReader struct {
	Dir     string
	Files   []*os.File
	Readers []*bufio.Reader
}

func NewFileListReader(dir string) (*DirReader, error) {
	d := &DirReader{
		Dir: dir,
	}
	err := d.scan()
	if err != nil {
		return nil, err
	}
	return d, nil
}

// how should I name this method? get()? look at dave cheney
func (f *DirReader) scan() error {

	filenames, err := f.buildFileList()
	if err != nil {
		return err
	}
	// Opening a file
	// Creating a Reader and reading the file line by line.
	// this basically acts as remove duplicates
	//print as list
	// We created a dict, key is ExitNode AKA node id, and we leverage files being ordered by date during
	// filepath.walk so that looping through all files should generate a map containing nodes with last update
	// for each node.
	f.Files = make([]*os.File, len(filenames))
	f.Readers = make([]*bufio.Reader, len(filenames))
	for i, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		// This should be safe because we use make to create a slice with len(filenames) capacity.
		// append does not work for some reason here. Study why.
		f.Readers[i] = bufio.NewReader(file)
		// We are going to use this list of files to close them all.
		f.Files[i] = file
		// test if this defer actually works.
	}
	return nil
}

func (f *DirReader) Close() {
	for _, file := range f.Files {
		// I guess nothing we can do if we get an error here...
		file.Close()
	}
}

// BuildFileList recursively walks inside a folder to generate the list of all
// files inside a folder tree. Items in the list comes out ordered.
func (f *DirReader) buildFileList() ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(f.Dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				fileList = append(fileList, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}
