package router

import (
	"github.com/daemonl/go_gsd/shared"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type router struct {
	routes             []*route
	parser             shared.IParser
	fallthroughHandler func(respWriter http.ResponseWriter, httpRequest *http.Request)
}

func GetRouter(p shared.IParser) Router {
	r := &router{}
	r.routes = make([]*route, 0, 0)
	r.parser = p
	return r
}

// AddRoute takes a format string as per the fmt and links it to a handler.
// The format string will be converted to a regular expression, currently only for %d and %s
// If any methods are specified, this path only applied to that method, default is all
func (r *router) AddRoute(format string, handler shared.IHandler, methods ...string) error {
	route, err := r.getRoute(format, handler, methods...)
	if err != nil {
		return err
	}
	r.routes = append(r.routes, route)
	return nil
}

func (r *router) getRoute(format string, handler shared.IHandler, methods ...string) (*route, error) {
	// Step 1, convert to a regexp.
	reStr := format
	reStr = strings.Replace(reStr, "%d", "[0-9]+", -1)
	reStr = strings.Replace(reStr, "%s", "[0-9A-Za-z_]+", -1)

	re, err := regexp.Compile("^" + reStr + "$")
	if err != nil {
		return nil, err
	}

	methodsUpper := make([]string, len(methods), len(methods))
	for i, m := range methods {
		methodsUpper[i] = strings.ToUpper(m)
	}
	nr := &route{
		format:  format,
		re:      re,
		handler: handler,
		methods: methodsUpper,
	}
	return nr, nil

}

func (r *router) AddRouteFunc(format string, hf func(shared.IRequest) (shared.IResponse, error), methods ...string) error {
	handler := handlerFunc(hf)
	return r.AddRoute(format, handler, methods...)
}

func (r *router) AddRoutePathFunc(format string, pathRequestFunc func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error {
	route, err := r.getRoute(format, nil, methods...)
	if err != nil {
		return err
	}
	normalRequestFunc := func(req shared.IRequest) (shared.IResponse, error) {
		pathRequest := wrapRequest(req, route)
		return pathRequestFunc(pathRequest)
	}
	handler := handlerFunc(normalRequestFunc)
	route.handler = handler
	return nil
}

func (r *router) getPathMatching(pathString string, method string) *route {
	var found *route

	method = strings.ToUpper(method)

searching:
	for _, p := range r.routes {
		if p.re.MatchString(pathString) {
			if len(p.methods) < 1 {
				found = p
				break searching
			}
			for _, m := range p.methods {
				if m == method {
					found = p
					break searching
				}

			}

		}
	}

	return found
}

func (r *router) Fallthrough(h func(respWriter http.ResponseWriter, httpRequest *http.Request)) {
	r.fallthroughHandler = h
}

func (r *router) ServeHTTP(respWriter http.ResponseWriter, httpRequest *http.Request) {

	var err error

	path := r.getPathMatching(httpRequest.URL.Path, httpRequest.Method)
	if path == nil {
		r.fallthroughHandler(respWriter, httpRequest)
		//log.Printf("Path '%s %s' did not match any\n", httpRequest.Method, httpRequest.URL.Path)
		//http.NotFound(respWriter, httpRequest)
		return
	}

	log.Printf("Path %s Matches %s\n", httpRequest.URL.Path, path.format)

	req, err := r.parser.Parse(respWriter, httpRequest)
	if err != nil {
		log.Printf("Parser error: %s\n", err.Error())
		respWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.Logf("Begin %s %s", httpRequest.Method, httpRequest.URL.RequestURI())
	req.Logf("User Agent: %s", httpRequest.UserAgent())
	defer req.Logf("End")

	res, err := path.handler.Handle(wrapRequest(req, path))
	if err != nil {
		req.Logf("ERROR: %s", err.Error())
		respWriter.WriteHeader(500)
		respWriter.Write([]byte(`INTERNAL SERVER ERROR`))
		return
	}
	if res == nil {
		return
	}

	contentType := res.ContentType()
	respWriter.Header().Add("Content-Type", contentType)

	err = res.WriteTo(respWriter)
	if err != nil {
		req.Logf("ERROR: %s", err.Error())
		respWriter.WriteHeader(500)
		respWriter.Write([]byte(`INTERNAL SERVER ERROR`))
		return
	}
}
