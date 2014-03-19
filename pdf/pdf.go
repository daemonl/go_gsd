package pdf

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
)

func DoPdf(binary string, in io.Reader, out io.Writer) error {

	br := bufio.NewReader(in)
	header := bytes.Buffer{}
	htmlStart := ([]byte("<"))[0]
	for {
		b, err := br.ReadByte()
		if err != nil {
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
			log.Printf("Error reading pdf header: %s\n", err)
			break
		}
		//line = line[:len(line)-1]
		log.Println(line)
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
	err := cmd.Run()

	ooString := outErr.String()
	if len(ooString) > 0 {
		log.Println("PDF ERROR: " + ooString)
		return errors.New(ooString)
	}
	if err != nil {
		log.Println("PDF ERROR: " + err.Error())
		return err
	}
	return nil
}
