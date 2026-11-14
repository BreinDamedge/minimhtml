package main

import (
	"fmt"
	"minimhtml.com/m/v2/minimhtml"
)

func main() {
	if err := minimhtml.Mht2Html("example_doc.mht", "example_doc.html"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success!")
	}
}
