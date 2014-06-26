package shared

import (
	"io"
	"net/http"
)

type IResponse interface {
	WriteTo(w io.Writer) error
	ContentType() string
	HTTPExtra(http.ResponseWriter)
}
