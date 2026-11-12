package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"os"
	"strings"
	"sync"
)

type MultipartSegment struct {
	Type             string
	TransferEncoding string
	Location         string
	Id               string
	Content          string
}

func MakeURI(segment *MultipartSegment) (string, error) {
	// assumes content is already decoded to utf-8 and is base64 encoded text
	return fmt.Sprintf("data:%s;charset=utf-8;base64,%s", segment.Type, segment.Content), nil
}

func decodeMultipartContent(segment *MultipartSegment) error {
	if segment.TransferEncoding == "quoted-printable" {
		r := quotedprintable.NewReader(strings.NewReader(segment.Content))
		decoded, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		segment.Content = string(decoded)
	} else if segment.TransferEncoding == "base64" {
		replacer := strings.NewReplacer("\r\n", "", "\n", "")
		clean := replacer.Replace(segment.Content)
		segment.Content = clean
	}
	return nil
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
	var documentSections []*MultipartSegment = []*MultipartSegment{}
	var wg sync.WaitGroup // for goroutining the decoding of the html segment
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
				header, value, found := strings.Cut(line, ":")
				if !found {
					// raise an error here
				}
				header = strings.TrimSpace(header)
				value = strings.TrimSpace(value)
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
				documentSections = append(documentSections, &section)
				wg.Go(func() {
					decodeMultipartContent(&section)
				})
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

	// wait for segments to be decoded so you can make the uris (for old new pairs)
	// TODO: you might be able to redo this with decoded flags so that you don't have to wait for just the last one to be done or smth.
	wg.Wait()

	// for each link in the html find it and replace it
	// use a replacer lol
	// make the uris and pair them with the original links
	oldNews := []string{}
	for _, segment := range documentSections {
		if segment.Type == "image/png" {
			oldNews = append(oldNews, segment.Location)
			uri, err := MakeURI(segment)
			if err != nil {
				log.Fatal(err)
			}
			oldNews = append(oldNews, uri)
		}
	}
	r := strings.NewReplacer(oldNews...)
	new_content := r.Replace(documentSections[0].Content)

	err = os.WriteFile("example_doc.html", []byte(new_content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
