package pdf

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
)

func DoPdf(binary string, in io.Reader, out io.Writer) error {
	log.Println("PDF")
	cmd := exec.Command(binary, "-q", "-", "-")
	var outErr bytes.Buffer
	cmd.Stdin = in
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
