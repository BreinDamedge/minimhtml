package main

import (
	"fmt"
	"github.com/breindamedge/minimhtml"
	"time"
)

func main() {
	start := time.Now() // Start time
	for range 1000 {
		minimhtml.Mht2Html("example_doc.mht")
	}
	duration := time.Since(start)
	fmt.Println(duration)
}
