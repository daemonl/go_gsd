package router

import (
	"database/sql"
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"io"
	"net/http"
)

type Handler interface {
	Handle(Request) (Response, error)
}

type Response interface {
	WriteTo(w io.Writer) error
	ContentType() string
	HTTPExtra(http.ResponseWriter)
}

type Request interface {
	Cleanup()
	Session() torch.Session
	IsLoggedIn() bool
	ResetSession() error
	Method() string
	Redirect(to string)
	DB() (*sql.DB, error)
	GetContext() torch.Context
	URLMatch(dest ...interface{}) error
	DoError(err error)
	DoErrorf(format string, parameters ...interface{})

	ScanPath(recievers ...interface{}) error

	GetRaw() (http.ResponseWriter, *http.Request)

	WriteString(content string)
	Writef(format string, params ...interface{})
	PostValueString(name string) string

	Broadcast(name string, val interface{})

	Write(bytes []byte) (int, error)

	Logf(string, ...interface{})
}

type Router interface {
	AddRoute(format string, handler Handler, methods ...string) error
	Fallthrough(func(respWriter http.ResponseWriter, httpRequest *http.Request))
	ServeHTTP(respWriter http.ResponseWriter, httpRequest *http.Request)
}

type Parser interface {
	Parse(http.ResponseWriter, *http.Request) (torch.Request, error)
}

type TorchHandlerFunc func(torch.Request) (torch.Response, error)

func (f TorchHandlerFunc) Handle(r Request) (Response, error) {
	return f(torch.Request(r))
}

type WrappedRequest struct {
	torch.Request
	route *route
}

func wrapRequest(tr torch.Request, route *route) Request {
	return &WrappedRequest{
		Request: tr,
		route:   route,
	}
}

func (wr *WrappedRequest) ScanPath(dests ...interface{}) error {
	_, r := wr.GetRaw()
	_, err := fmt.Sscanf(r.URL.Path, wr.route.format, dests...)
	return err
}
