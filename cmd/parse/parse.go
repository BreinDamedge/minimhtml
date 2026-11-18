package main

import (
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
		out, err := minimhtml.Mht2Html(path)
		// write the file
		data := []byte(out)

		// 0644 sets standard read/write permissions for the owner
		err = os.WriteFile(path[:len(path)-3]+"html", data, 0644)
		if err != nil {
			panic(err)
		}
	}
}
