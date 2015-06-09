package router

import (
	"github.com/daemonl/go_gsd/shared"
	//"github.com/daemonl/go_gsd/torch"
	"net/http"
)

type Router interface {
	AddRoute(format string, handler shared.IHandler, methods ...string) error
	AddRouteFunc(format string, handlerFunc func(shared.IRequest) (shared.IResponse, error), methods ...string) error
	AddRoutePathFunc(format string, handlerPathFunc func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error
	Fallthrough(string)
	ServeHTTP(respWriter http.ResponseWriter, httpRequest *http.Request)
	Redirect(from, to string)
}

type UserDisplayError interface {
	Error() string
	GetUserDescription() string
	GetHTTPStatus() int
}

type UserObjectError interface {
	Error() string
	GetHTTPStatus() int
	GetUserObject() interface{}
}

/*
type TorchHandlerFunc func(torch.Request) (torch.Response, error)

func (f TorchHandlerFunc) Handle(r Request) (Response, error) {
	return f(torch.Request(r))
}

type IRequest interface {
	shared.IRequest
}
*/

type handlerFunc func(shared.IRequest) (shared.IResponse, error)

func (fn handlerFunc) Handle(req shared.IRequest) (shared.IResponse, error) {
	return fn(req)
}
