package router

import (
	"github.com/daemonl/go_gsd/torch"
	"io"
	"net/http"
)

type Handler interface {
	Handle(torch.Request) (Response, error)
}

type Response interface {
	WriteTo(w io.Writer) error
	ContentType() string
}

type Router interface {
	AddRoute(format string, handler Handler, methods ...string) error
	Fallthrough(func(respWriter http.ResponseWriter, httpRequest *http.Request))
	ServeHTTP(respWriter http.ResponseWriter, httpRequest *http.Request)
}

type Parser interface {
	Parse(http.ResponseWriter, *http.Request) (torch.Request, error)
}
