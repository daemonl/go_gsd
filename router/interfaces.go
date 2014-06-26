package router

import (
	"fmt"
	"github.com/daemonl/go_gsd/shared"
	//"github.com/daemonl/go_gsd/torch"
	"net/http"
)

type Router interface {
	AddRoute(format string, handler shared.IHandler, methods ...string) error
	AddRouteFunc(format string, handlerFunc func(shared.IRequest) (shared.IResponse, error), methods ...string) error
	AddRoutePathFunc(format string, handlerPathFunc func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error
	Fallthrough(func(respWriter http.ResponseWriter, httpRequest *http.Request))
	ServeHTTP(respWriter http.ResponseWriter, httpRequest *http.Request)
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

type WrappedRequest struct {
	shared.IRequest
	route *route
}

func wrapRequest(tr shared.IRequest, route *route) shared.IPathRequest {
	return &WrappedRequest{
		IRequest: tr,
		route:    route,
	}
}

func (wr *WrappedRequest) ScanPath(dests ...interface{}) error {
	_, r := wr.GetRaw()
	_, err := fmt.Sscanf(r.URL.Path, wr.route.format, dests...)
	return err
}
