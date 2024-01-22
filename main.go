package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)


// TODO: there is no need for a temp file
// need to find a way to extract properly all these files
// a final cleanup of all text files must be done


func main() {
	downloadFile("https://collector.torproject.org/archive/exit-lists/exit-list-2024-01.tar.xz")
}

func downloadFile(uri string) {

	fileName := path.Base(uri)
	file, err := os.CreateTemp("", fileName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(file.Name())

	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		panic(err)
	}

}
