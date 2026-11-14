package main

import (
	"fmt"
	"github.com/breindamedge/minimhtml"
)

func main() {
	if _, err := minimhtml.Mht2Html("example_doc.mht"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success!")
	}
}
