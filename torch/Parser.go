package torch

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
)

type Parser struct {
	Store                  SessionStore
	DB                     *sql.DB
	OpenDatabaseConnection func(session *Session) (*sql.DB, error)
	PublicPatterns         []*regexp.Regexp
}

// Wraps a function expecting a Request to make it work with httpResponseWriter, http.Request
func (parser *Parser) Wrap(handler func(Request)) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// Verbose request logging
		log.Printf("Begin Request %s\n", r.RequestURI)
		d, _ := httputil.DumpRequest(r, false)
		log.Println(string(d))
		defer log.Printf("End Request\n")

		requestTorch, err := parser.parseRequest(w, r)
		if err != nil {
			log.Fatal(err)
			w.Write([]byte("An error occurred"))
			return
		}
		defer requestTorch.Cleanup()

		/*
			db, err := parser.OpenDatabaseConnection(requestTorch.Session)
			if err != nil {
				log.Fatal(err)
				w.Write([]byte("An error occurred"))
				return
			}
			requestTorch.db = db
			//defer requestTorch.db.Close()
		*/

		if !requestTorch.IsLoggedIn() {
			log.Printf("PUBLIC: Check Path %s", r.URL.Path)
			for _, p := range parser.PublicPatterns {
				log.Println(p.String())
				if p.MatchString(r.URL.Path) {
					log.Printf("PUBLIC: Matched Public Path %s", p.String())
					handler(requestTorch)
					return
				}
			}

			log.Println("PUBLIC: No Public Pathes Matched")

			requestTorch.Redirect("/login")
		} else {
			handler(requestTorch)
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
func (parser *Parser) WrapSplit(handlers ...func(Request)) func(w http.ResponseWriter, r *http.Request) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return parser.Wrap(func(request Request) {
		for i, m := range methods {
			if request.Method() == m {
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
// ParseRequest opens a database session for request.DB(), It will need to be closed...
func (parser *Parser) parseRequest(w http.ResponseWriter, r *http.Request) (Request, error) {
	request := basicRequest{
		writer: w,
		raw:    r,
	}

	sessCookie, err := r.Cookie("gsd_session")
	if err != nil {
		request.ResetSession()
	} else {
		sess, err := parser.Store.GetSession(sessCookie.Value)
		if err != nil {
			request.ResetSession()
		} else {
			request.session = sess
		}
	}

	return &request, nil
}
