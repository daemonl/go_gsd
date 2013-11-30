package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	in := strings.NewReader("Hello World")
	var out bytes.Buffer
	DoPdf("/Applications/wkhtmltopdf.app/Contents/MacOS/wkhtmltopdf", in, &out)
	fmt.Println("before")
	fmt.Println(out.String())
	fmt.Println("done")
}
