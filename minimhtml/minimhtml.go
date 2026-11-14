package minimhtml

import (
	"bufio"
	"encoding/base64"
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

func base64Encode(segment *MultipartSegment) (string, error) {
	switch segment.TransferEncoding {
	case "base64":
		return segment.Content, nil
	case "quoted-printable":
		r := quotedprintable.NewReader(strings.NewReader(segment.Content))
		decoded, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(decoded))
		return encoded, nil
	default:
		return "", fmt.Errorf("Unsupported Transfer-Encoding: %s\n", segment.TransferEncoding)
	}
}

func MakeURI(segment *MultipartSegment) (string, error) {
	switch segment.Type {
	case "image/png":
		// assumes content is already decoded from quoted-printable and is base64 encoded text
		return fmt.Sprintf("data:%s;charset=utf-8;base64,%s", segment.Type, segment.Content), nil
	case "text/css":
		b64Payload, err := base64Encode(segment)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("data:%s;charset=utf-8;base64,%s", segment.Type, b64Payload), nil
	default:
		return "", fmt.Errorf("Unsupported Content-Type for URI: %s\n", segment.Type)
	}
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

func Mht2Html(src string, dst string) error {

	// open the file
	file, err := os.Open(src)
	if err != nil {
		return err
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
	var wg sync.WaitGroup // for goroutining the decoding of the html segment  TODO: consider instead using a "decoded" flag so that you don't actually do any waiting for a piece of data unless you're trying to use it.
	// the final boundary of a multipart body is the boundary line followed by "--" (RFC 2046),
	// so both forms have to be treated as section delimiters
	finalSep := sep + "--"

	// now collect the sections
	for {
		// you've hit a new section or the end of the file so make a new MultipartSegment to keep track of this info just in case
		section := MultipartSegment{}
		sawHeader := false
		// figure out what type of chunk this is (read the headers)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" { // if you hit an empty line under headers then you're about to hit content so stop the header reading
				break
			} else {
				sawHeader = true
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

		// NOTE:
		// no more sections (we hit EOF while looking for headers)
		// if there were no more sections, the header-reading loop will have
		// hit EOF, sawHeader will stay false, and we break the outer loop
		if !sawHeader {
			break
		}

		// read the content of the section until a boundary (or EOF, for the last segment)
		var br strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if line == sep || line == finalSep { // if you hit the separator you've reached a new section or the end of the body
				break
			}
			// consume body
			br.WriteString(line)
			br.WriteString("\n") // add the newline back since the scanner removes it
		}

		// put this section into the list and start on the next segment
		// (this also catches the last segment, which ends at EOF instead of a separator)
		section.Content = br.String()
		documentSections = append(documentSections, &section)
		wg.Go(func() {
			decodeMultipartContent(&section)
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// write the content of the html blob into an html file
	if len(documentSections) == 0 {
		log.Fatal(fmt.Errorf("documentSections cannot be empty: length is %d\n", len(documentSections)))
	}

	// wait for segments to be decoded so you can make the uris (for old new pairs)
	// TODO: you might be able to redo this with decoded flags so that you don't have to wait for just the last one to be done or smth when you're not trying to use that section yet.
	wg.Wait()

	// for each link in the html find it and replace it
	// use a replacer lol
	// make the uris and pair them with the original links
	oldNews := []string{}
	for _, segment := range documentSections {
		if segment.Type == "image/png" || segment.Type == "text/css" {
			oldNews = append(oldNews, segment.Location)
			uri, err := MakeURI(segment)
			if err != nil {
				return err
			}
			oldNews = append(oldNews, uri)
		}
	}

	r := strings.NewReplacer(oldNews...)
	new_content := r.Replace(documentSections[0].Content)

	err = os.WriteFile(dst, []byte(new_content), 0644)
	if err != nil {
		return err
	}

	return nil
}
