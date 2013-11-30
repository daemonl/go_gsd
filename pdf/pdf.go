package pdf

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

func DoPdf(binary string, in io.Reader, out io.Writer) error {
	cmd := exec.Command(binary, "-q", "-", "-")
	var outErr bytes.Buffer
	cmd.Stdin = in
	cmd.Stderr = &outErr
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return err
	}
	ooString := outErr.String()
	if len(ooString) > 0 {
		return errors.New(ooString)
	}
	return nil
}
