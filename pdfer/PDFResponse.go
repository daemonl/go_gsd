package pdfer

import (
	"bufio"
	"bytes"
	"io"

	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/daemonl/go_gsd/shared"
)

type PDFResponse struct {
	binary string
	htmlIn shared.IResponse
}

func (r *PDFResponse) ContentType() string {
	return "application/pdf"

}
func (r *PDFResponse) WriteTo(out io.Writer) error {

	binary := r.binary

	bufferedResponse := &bytes.Buffer{}
	err := r.htmlIn.WriteTo(bufferedResponse)
	if err != nil {
		return fmt.Errorf("writing PDF response: %s", err.Error())
	}

	br := bufio.NewReader(bufferedResponse)

	header := bytes.Buffer{}
	htmlStart := ([]byte("<"))[0]
	for {
		b, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				log.Printf("HEADER: %s\n", header.String())
				break
			}
			log.Printf("Error reading pdf header: %s\n", err)
			break
		}
		if b == htmlStart {
			br.UnreadByte()
			break
		} else {
			header.WriteByte(b)
		}
	}

	parameters := []string{
		"-q",
	}
	headerReader := bufio.NewReader(&header)
	for {
		line, err := headerReader.ReadString('\n')
		if err != nil {
			log.Printf("Error parsing pdf header: %s\n", err)
			break
		}
		//line = line[:len(line)-1]
		log.Println(line)
		if !strings.HasPrefix(line, "-") {
			log.Printf("Discarding non flag line: %s\n", line)
			break
		}
		parts := strings.SplitN(line, " ", 2)
		parameters = append(parameters, strings.TrimSpace(parts[0]))
		if len(parts) == 2 {
			parameters = append(parameters, strings.TrimSpace(parts[1]))
		}
	}

	log.Printf("%#v\n", parameters)

	parameters = append(parameters, "-", "-")
	cmd := exec.Command(binary, parameters...)
	var outErr bytes.Buffer
	cmd.Stdin = br
	cmd.Stderr = &outErr
	cmd.Stdout = out
	err = cmd.Run()

	ooString := outErr.String()
	if len(ooString) > 0 {
		return fmt.Errorf(ooString)
	}
	if err != nil {
		log.Println("PDF ERROR: " + err.Error())
		return err
	}
	return nil
}

func (r *PDFResponse) HTTPExtra(w http.ResponseWriter) {

}
