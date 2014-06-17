package torch

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/daemonl/databath"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Request struct {
	writer  http.ResponseWriter
	raw     *http.Request
	db      *sql.DB
	Session *Session
	Method  string
}

func (r *Request) DB() (*sql.DB, error) {
	return r.db, nil
}

func (r *Request) GetWriter() http.ResponseWriter {
	return r.writer
}

func (r *Request) GetRaw() (http.ResponseWriter, *http.Request) {
	return r.writer, r.raw
}

func (r *Request) UrlMatch(dest ...interface{}) error {
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
func (r *Request) DoError(err error) {
	log.Println(err)
	r.Writef("Error: %s", err.Error())
}

func (r *Request) DoErrorf(format string, parameters ...interface{}) {
	str := fmt.Sprintf(format, parameters...)
	log.Println(str)
	r.Write(str)
}

func (r *Request) GetSession() *Session {
	return r.Session
}

func (r *Request) Broadcast(name string, val interface{}) {
	r.Session.Store.Broadcast(name, val)

}

func (r *Request) GetContext() databath.Context {
	context := &databath.MapContext{
		IsApplication:   false,
		UserAccessLevel: r.Session.User.Access,
		Fields:          make(map[string]interface{}),
	}
	context.Fields["me"] = r.Session.User.Id
	context.Fields["user"] = r.Session.User.Id
	return context
}
