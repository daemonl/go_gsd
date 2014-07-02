package router

import (
	"fmt"
	"github.com/daemonl/go_gsd/shared"
	//"github.com/daemonl/go_gsd/torch"
	"net/http"
	"net/url"
	"strings"
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

	uri := strings.Replace(r.URL.RequestURI(), "/", " ", -1)
	format := strings.Replace(wr.route.format, "/", " ", -1)
	_, err := fmt.Sscanf(uri, format, dests...)
	for _, d := range dests {
		switch d := d.(type) {
		case *string:
			newVal, err := url.QueryUnescape(*d)
			if err != nil {
				return err
			}
			*d = newVal
		}
	}
	if err != nil {
		return fmt.Errorf("scanning '%s' into '%s': %s", r.URL.Path, wr.route.format, err.Error())
	}

	return err
}
