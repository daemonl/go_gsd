package shared

import (
	"fmt"
	"io"
	"net/http"
	"encoding/json"
)

type IResponse interface {
	WriteTo(w io.Writer) error
	ContentType() string
	HTTPExtra(http.ResponseWriter)
}

type quickStringResponse struct {
	msg string
}

func (qr *quickStringResponse) WriteTo(w io.Writer) error {
	fmt.Fprint(w, qr.msg)
	return nil
}

func (qr *quickStringResponse) ContentType() string             { return "text/plain" }
func (qr *quickStringResponse) HTTPExtra(r http.ResponseWriter) {}
func QuickStringResponse(msg string) IResponse {
	return &quickStringResponse{msg: msg}
}

type quickJSONResponse struct {
	obj interface{}
}

func (qr *quickJSONResponse) WriteTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(qr.obj)
	return err
}

func (qr *quickJSONResponse) ContentType() string             { return "application/json" }
func (qr *quickJSONResponse) HTTPExtra(r http.ResponseWriter) {}
func QuickJSONResponse(obj interface{}) IResponse {
	return &quickJSONResponse{obj: obj}
}
