package shared

import (
	"net/http"
)

type IParser interface {
	Parse(http.ResponseWriter, *http.Request) (IRequest, error)
}
