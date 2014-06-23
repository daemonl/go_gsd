package torch

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type basicRequest struct {
	writer  http.ResponseWriter
	raw     *http.Request
	db      *sql.DB
	session Session
}

func (r *basicRequest) DB() (*sql.DB, error) {
	return r.db, nil
}

func (r *basicRequest) Method() string {
	return r.raw.Method
}

func (r *basicRequest) Cleanup() {

}

func (r *basicRequest) Session() Session {
	return r.session
}

func (r *basicRequest) IsLoggedIn() bool {
	if r.session == nil {
		return false
	}
	if r.session.User() == nil {
		return false
	}
	return true
}

func (r *basicRequest) GetWriter() http.ResponseWriter {
	return r.writer
}

func (r *basicRequest) GetRaw() (http.ResponseWriter, *http.Request) {
	return r.writer, r.raw
}

func (r *basicRequest) UrlMatch(dest ...interface{}) error {
	urlParts := strings.Split(r.raw.URL.Path[1:], "/")
	if len(urlParts) != len(dest) {
		fmt.Println(urlParts)
		return errors.New(fmt.Sprintf("URL had %d parameters, expected %d", len(urlParts), len(dest)))
	}
	for i, src := range urlParts {
		dst := dest[i]
		switch t := dst.(type) {
		case *string:
			*t = src
		case *uint64:
			srcInt, err := strconv.ParseUint(src, 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("URL Parameter %d could not be converted to an unsigned integer"))
			}
			*t = srcInt

		default:
			return errors.New(fmt.Sprintf("URL Parameter %d could not be converted to a %T",
				i+1, t))

		}
	}
	return nil
}
func (r *basicRequest) DoError(err error) {
	log.Println(err)
	r.Writef("Error: %s", err.Error())
}

func (r *basicRequest) DoErrorf(format string, parameters ...interface{}) {
	str := fmt.Sprintf(format, parameters...)
	log.Println(str)
	r.Write(str)
}

func (r *basicRequest) Broadcast(name string, val interface{}) {
	r.Session().Broadcast(name, val)
}

func (request *basicRequest) End() {

}

func (request *basicRequest) Write(content string) {
	request.writer.Write([]byte(content))
}

func (request *basicRequest) Writef(format string, params ...interface{}) {
	request.Write(fmt.Sprintf(format, params...))
}

func (request *basicRequest) PostValueString(name string) string {
	return request.raw.PostFormValue(name)
}

func (request *basicRequest) Redirect(to string) {
	http.Redirect(request.writer, request.raw, to, 302)
}

func (request *basicRequest) ResetSession() error {
	s, err := request.session.SessionStore().NewSession()
	if err != nil {
		return err
	}
	request.session = s
	c := http.Cookie{Name: "gsd_session", Path: "/", MaxAge: 86400, Value: *request.session.Key()}
	request.writer.Header().Add("Set-Cookie", c.String())
	return nil
}
