package torch

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/daemonl/go_gsd/shared"
)

type basicRequest struct {
	writer  http.ResponseWriter
	raw     *http.Request
	db      *sql.DB
	session shared.ISession
}

func (r *basicRequest) DB() (*sql.DB, error) {
	if r.db == nil {
		db, err := r.session.GetDatabaseConnection()
		if err != nil {
			return nil, err
		}
		r.db = db
	}
	return r.db, nil
}

func (r *basicRequest) QueryString() shared.IQueryString {
	return shared.GetQueryString(r.raw.URL.Query())
}

func (r *basicRequest) Method() string {
	return r.raw.Method
}

func (r *basicRequest) Cleanup() {
	if r.db == nil {
		return
	}
	r.session.ReleaseDB(r.db)
	r.db = nil

}

func (r *basicRequest) Session() shared.ISession {
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

func (r *basicRequest) GetContext() shared.IContext {
	if r.session == nil {
		return nil
	}
	u := r.session.User()
	if u == nil {
		return nil
	}
	return u.GetContext()
}

func (r *basicRequest) GetRaw() (http.ResponseWriter, *http.Request) {
	return r.writer, r.raw
}

func (r *basicRequest) ReadJson(into interface{}) error {
	dec := json.NewDecoder(r.raw.Body)
	return dec.Decode(into)
}

func (r *basicRequest) DoError(err error) {
	log.Println(err)
	r.Writef("Error: %s", err.Error())
}

func (r *basicRequest) DoErrorf(format string, parameters ...interface{}) {
	str := fmt.Sprintf(format, parameters...)
	log.Println(str)
	r.WriteString(str)
}

func (r *basicRequest) Broadcast(name string, val interface{}) {
	r.Session().Broadcast(name, val)
}

func (request *basicRequest) End() {

}

func (request *basicRequest) Write(bytes []byte) (int, error) {
	return request.writer.Write(bytes)
}

func (request *basicRequest) WriteString(content string) {
	request.writer.Write([]byte(content))
}

func (request *basicRequest) Writef(format string, params ...interface{}) {
	request.WriteString(fmt.Sprintf(format, params...))
}

func (request *basicRequest) PostValueString(name string) string {
	return request.raw.PostFormValue(name)
}

func (request *basicRequest) Redirect(to string) {
	http.Redirect(request.writer, request.raw, to, 302)
}

func (request *basicRequest) ResetSession() error {
	if request.session == nil {
		return fmt.Errorf("No session when reset session was called")
	}
	s, err := request.session.SessionStore().NewSession()
	if err != nil {
		return err
	}
	request.SetSession(s)

	return nil
}

func (request *basicRequest) SetSession(session shared.ISession) {
	request.session = session
	c := http.Cookie{Name: "gsd_session", Path: "/", MaxAge: 86400, Value: *request.session.Key()}
	request.writer.Header().Add("Set-Cookie", c.String())
}

func (request *basicRequest) Logf(format string, params ...interface{}) {
	log.Printf(format, params...)
}
