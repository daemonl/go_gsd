package torch

import (
	"github.com/daemonl/go_lib/databath"
	"log"
	"net/http"
	"regexp"
)

type Parser struct {
	Store          *SessionStore
	Bath           *databath.Bath
	PublicPatterns []*regexp.Regexp
}

type Request struct {
	writer  http.ResponseWriter
	raw     *http.Request
	DbConn  *databath.Connection
	Session *Session
	Method  string
}

func (r *Request) GetWriter() http.ResponseWriter {
	return r.writer
}
func (r *Request) GetRaw() (http.ResponseWriter, *http.Request) {
	return r.writer, r.raw
}

// Wraps a function expecting a Request to make it work with httpResponseWriter, http.Request
func (parser *Parser) WrapReturn(handler func(*Request)) func(w http.ResponseWriter, r *http.Request) *Request {
	return func(w http.ResponseWriter, r *http.Request) *Request {

		requestTorch, err := parser.ParseRequest(w, r)
		if err != nil {
			log.Fatal(err)
			w.Write([]byte("An error occurred"))
			return nil
		}
		if requestTorch.Session.User == nil {
			log.Println("NOT LOGGED IN")
			log.Printf("Check Path %s", r.URL.Path)
			for _, p := range parser.PublicPatterns {
				log.Println(p.String())
				if p.MatchString(r.URL.Path) {
					log.Printf("Matched Public Path %s", p.String())
					handler(requestTorch)
					return requestTorch
				}
			}
			log.Println("No Public Pathes Matched")
			requestTorch.Redirect("/login")
		} else {
			handler(requestTorch)
		}
		return requestTorch
	}
}

func (parser *Parser) Wrap(handler func(*Request)) func(w http.ResponseWriter, r *http.Request) {
	f := parser.WrapReturn(handler)

	return func(w http.ResponseWriter, r *http.Request) {
		_ = f(w, r)
		return
	}
}

// WrapSplit checks the method of a request, and uses the handlers passed in order GET, POST, PUT, DELETE.
// To skip a method, pass nil (Or don't specify) Will return 404
// The order is an obsuciry I'm not proud of... probably should be a map?
// Ideally the methods should be registered separately, but that requires taking over more of the default
// functionality in httpRequestHandler which is not the plan at this stage. - These are Helpers, not a framework.
// (Who am I kidding, I love building frameworks)
func (parser *Parser) WrapSplit(handlers ...func(*Request)) func(w http.ResponseWriter, r *http.Request) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return parser.Wrap(func(request *Request) {
		for i, m := range methods {
			if request.Method == m {
				if len(handlers) > i && handlers[i] != nil {
					handlers[i](request)
				} else {
					//TODO: 404
				}
				return
			}
		}
	})
}

// ParseRequest is a utility usually used internally to give a Request object to a standard http request
// Exported for better flexibility
func (parser *Parser) ParseRequest(w http.ResponseWriter, r *http.Request) (*Request, error) {
	request := Request{
		writer: w,
		raw:    r,
		Method: r.Method,
		DbConn: parser.Bath.GetConnection(),
	}
	defer request.DbConn.Release()

	sessCookie, err := r.Cookie("gsd_session")
	if err != nil {
		request.NewSession(parser.Store)
	} else {
		sess, err := parser.Store.GetSession(sessCookie.Value)
		if err != nil {
			request.NewSession(parser.Store)
		} else {
			request.Session = sess
		}
	}

	return &request, nil
}
