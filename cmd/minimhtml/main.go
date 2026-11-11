package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"os"
	"strings"
)

type MultipartSegment struct {
	Type             string
	TransferEncoding string
	Location         string
	Id               string
	Content          string
}

func main() {
	// open the file
	file, err := os.Open("example_doc.mht")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// setup stuff for identifying the separator
	sep_prefix := "------MultipartBoundary--"
	var sep string

	// start reading and find the separator
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, sep_prefix) {
			sep = line
			break
		}
	}

	// make a place to store info & data of sections you find while parsing
	var documentSections []MultipartSegment = []MultipartSegment{}

	// now collect the sections
	for {
		// you've hit a new section or the end of the file so make a new MultipartSegment to keep track of this info just in case
		section := MultipartSegment{}
		done := true
		// figure out what type of chunk this is (read the headers)
		for scanner.Scan() {
			// if scan succeeds set done to false
			done = false
			line := scanner.Text()
			if line == "" { // if you hit an empty line under headers then you're about to hit content so stop the header reading
				break
			} else {
				// parse headers
				parts := strings.Split(line, ":")
				header := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				switch header {
				case "Content-Type":
					section.Type = value
				case "Content-Transfer-Encoding":
					section.TransferEncoding = value
				case "Content-Location":
					section.Location = value
				case "Content-ID":
					section.Id = value
				default:
					fmt.Printf("Unknown header: %s\n", header)
				}
			}
		}

		// if read failed, done will be false
		if done {
			break
		}

		// read the content of the section
		var br strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if line == sep { // if you hit the separator you've reached a new section
				// put this section into the list and start on the next segment
				section.Content = br.String()
				documentSections = append(documentSections, section)
				break
			} else {
				// consume body
				br.WriteString(line)
				br.WriteString("\n") // add the newline back since the scanner removes it
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// write the content of the html blob into an html file
	if len(documentSections) == 0 {
		log.Fatal(fmt.Errorf("documentSections cannot be empty: length is %d\n", len(documentSections)))
	}
	r := quotedprintable.NewReader(strings.NewReader(documentSections[0].Content))
	decoded, err := io.ReadAll(r)

	err = os.WriteFile("example_doc.html", []byte(decoded), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
