package torch

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"

	"github.com/daemonl/go_gsd/shared"
)

type Parser struct {
	Store                  shared.ISessionStore
	OpenDatabaseConnection func(session shared.ISession) (*sql.DB, error)
	PublicPatterns         []*regexp.Regexp
}

func BasicParser(sessionStore shared.ISessionStore, rawPublicPatterns []string) *Parser {
	parser := &Parser{
		Store:          sessionStore,
		PublicPatterns: make([]*regexp.Regexp, len(rawPublicPatterns), len(rawPublicPatterns)),
	}

	// Insert the regexes for all 'public' urls
	for i, pattern := range rawPublicPatterns {
		reg := regexp.MustCompile(pattern)
		parser.PublicPatterns[i] = reg
	}

	return parser
}

// Wraps a function expecting a Request to make it work with httpResponseWriter, http.Request
func (parser *Parser) Wrap(handler func(shared.IRequest) error) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// Verbose request logging
		log.Printf("Begin Request %s\n", r.RequestURI)
		d, _ := httputil.DumpRequest(r, false)
		log.Println(string(d))
		defer log.Printf("End Request\n")

		request, err := parser.Parse(w, r)
		if err != nil {
			log.Fatal(err)
			w.Write([]byte("An error occurred"))
			return
		}
		defer request.Cleanup()

		if !request.IsLoggedIn() {
			log.Printf("PUBLIC: Check Path %s", r.URL.Path)
			for _, p := range parser.PublicPatterns {
				log.Println(p.String())
				if p.MatchString(r.URL.Path) {
					log.Printf("PUBLIC: Matched Public Path %s", p.String())
					handler(request)
					return
				}
			}

			log.Println("PUBLIC: No Public Pathes Matched")

			request.Redirect("/login")
		} else {
			err := handler(request)
			if err != nil {
				fmt.Printf("Handler Error: %s\n", err.Error())
				request.DoError(err)
			}
		}
		return
	}
}

// WrapSplit checks the method of a request, and uses the handlers passed in order GET, POST, PUT, DELETE.
// To skip a method, pass nil (Or don't specify) Will return 404
// The order is an obsuciry I'm not proud of... probably should be a map?
// Ideally the methods should be registered separately, but that requires taking over more of the default
// functionality in httpRequestHandler which is not the plan at this stage. - These are Helpers, not a framework.
// (Who am I kidding, I love building frameworks)
func (parser *Parser) WrapSplit(handlers ...func(shared.IRequest)) func(w http.ResponseWriter, r *http.Request) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return parser.Wrap(func(request shared.IRequest) error {
		for i, m := range methods {
			if request.Method() == m {
				if len(handlers) > i && handlers[i] != nil {
					handlers[i](request)
				} else {
					//TODO: 404
				}
				return nil
			}
		}
		return nil
	})
}

// ParseRequest is a utility usually used internally to give a Request object to a standard http request
// Exported for better flexibility
// ParseRequest opens a database session for request.DB(), It will need to be closed...
func (parser *Parser) Parse(w http.ResponseWriter, r *http.Request) (shared.IRequest, error) {
	request := basicRequest{
		writer: w,
		raw:    r,
	}

	sessCookie, err := r.Cookie("gsd_session")
	if err != nil {
		log.Printf("Error getting cookie: %s\n", err.Error())
		sess, err := parser.Store.NewSession()
		if err != nil {
			return nil, err
		}
		request.SetSession(sess)
	} else {
		//log.Printf("No err geting cookie: %s\n", sessCookie.Value)
		sess, err := parser.Store.GetSession(sessCookie.Value)
		if err != nil {
			log.Printf("Error getting session from cookie: %s\n", err.Error())
		}
		if sess == nil {
			log.Println("Session cookie not found")
			sess, err = parser.Store.NewSession()
			if err != nil {
				return nil, fmt.Errorf("Could not load session: %s", err.Error())
			}
			if sess == nil {
				return nil, fmt.Errorf("Could not load session: returned nill")
			}
		}
		request.SetSession(sess)
	}

	return &request, nil
}
