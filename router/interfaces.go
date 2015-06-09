package router

import (
	"fmt"
	"strconv"

	"github.com/daemonl/go_gsd/shared"
	//"github.com/daemonl/go_gsd/torch"
	"net/http"
	"net/url"
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

	urlParts := wr.route.re.FindStringSubmatch(r.URL.RequestURI())

	if len(urlParts) != len(dests)+1 {
		return fmt.Errorf("scanning '%s' into '%s', %d parts, length mismatch", r.URL.Path, len(dests), wr.route.re.String())
	}

	for idx, dest := range dests {
		raw, err := url.QueryUnescape(urlParts[idx+1])
		if err != nil {
			return fmt.Errorf("scanning '%s' into '%s': %s", r.URL.Path, wr.route.re.String(), err.Error())
		}
		switch d := dest.(type) {
		case *string:
			*d = raw
		case *uint64:
			num, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = num
		case *int64:
			num, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = num
		case *uint32:
			num, err := strconv.ParseUint(raw, 10, 32)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = uint32(num)
		case *int32:
			num, err := strconv.ParseInt(raw, 10, 32)
			if err != nil {
				return fmt.Errorf("Type conversion error from %s to number: %s", raw, err.Error())
			}
			*d = int32(num)
		default:
			return fmt.Errorf("Type %T not implemented for URL matching", d)
		}
	}

	return nil
}
