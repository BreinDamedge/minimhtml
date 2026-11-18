package main

import (
	"fmt"
	"github.com/breindamedge/minimhtml"
	"math/rand"
	"os"
	"time"
)

func main() {
	TEST_DIRECTORY := "./TestDocuments"

	// read the files from the test directory (get all their paths)
	doc_paths, err := os.ReadDir(TEST_DIRECTORY)
	if err != nil {
		panic(err)
	}

	fileBuffers := [][]byte{}
	for i := range doc_paths {
		// ignore non-mht files
		path := TEST_DIRECTORY + "/" + doc_paths[i].Name()
		if path[len(path)-3:] != "mht" { // check file extention
			continue
		}

		buff, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		fileBuffers = append(fileBuffers, buff)
	}

	numFiles := len(fileBuffers)
	// begin parsing
	start := time.Now() // Start time
	for range 1000 {
		// sample a doc at random
		i := rand.Int() % numFiles
		buff := fileBuffers[i]
		minimhtml.BuffMht2Html(buff)
	}
	duration := time.Since(start)
	fmt.Println(duration)
}
