package main

import (
	"fmt"
	"github.com/breindamedge/minimhtml"
	"os"
)

func main() {
	TEST_DIRECTORY := "./TestDocuments"

	// read the files from the test directory (get all their paths)
	doc_paths, err := os.ReadDir(TEST_DIRECTORY)
	if err != nil {
		panic(err)
	}

	for i := range doc_paths {
		path := TEST_DIRECTORY + "/" + doc_paths[i].Name()
		if path[len(path)-3:] != "mht" { // check file extention
			continue
		}

		// open file
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// read it's locations
		locations, err := minimhtml.ExtractLocations(file)
		if err != nil {
			fmt.Printf("error occured while processing %s\n", path)
			panic(err)
		}

		fmt.Printf("%s:\n", path)
		for i := range locations {
			fmt.Println(locations[i])
		}
		fmt.Println()
	}
}
