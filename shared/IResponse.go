package shared

import (
	"fmt"
	"io"
	"net/http"
)

type IResponse interface {
	WriteTo(w io.Writer) error
	ContentType() string
	HTTPExtra(http.ResponseWriter)
}

type quickStringResponse struct {
	msg string
}

func (qr *quickStringResponse) WriteTo(w io.Writer)  error           {
	fmt.Fprint(w, qr.msg)
return nil
}
func (qr *quickStringResponse) ContentType() string             { return "text/plain" }
func (qr *quickStringResponse) HTTPExtra(r http.ResponseWriter) {}
func QuickStringResponse(msg string) IResponse {
	return &quickStringResponse{msg: msg}
}
