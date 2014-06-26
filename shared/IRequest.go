package shared

import (
	"database/sql"
	"net/http"
)

type IRequest interface {
	Cleanup()
	Session() ISession
	IsLoggedIn() bool
	ResetSession() error
	Method() string
	Redirect(to string)
	DB() (*sql.DB, error)
	GetContext() IContext
	URLMatch(dest ...interface{}) error
	DoError(err error)
	DoErrorf(format string, parameters ...interface{})

	GetRaw() (http.ResponseWriter, *http.Request)

	WriteString(content string)
	Writef(format string, params ...interface{})
	PostValueString(name string) string

	Broadcast(name string, val interface{})

	Write(bytes []byte) (int, error)

	Logf(string, ...interface{})
}
