package torch

import (
	"database/sql"
	"io"
	"net/http"
	"time"
)

type LoginLogout interface {
	ForceLogin(request Request, email string)
	LoadUserById(id uint64) (User, error)
	HandleLogin(Request) (Response, error)
	HandleLogout(Request) (Response, error)
	HandleSetPassword(Request) (Response, error)
}

type SessionStore interface {
	GetSession(key string) (Session, error)
	NewSession() (Session, error)
	DumpSessions()
	SetBroadcast(func(string, interface{}))
	GetDatabaseConnectionForSession(Session) (*sql.DB, error)
}

type User interface {
	GetContext() Context
	CheckPassword(string) (bool, error)
	ID() uint64
	Access() uint64
	WhoAmIObject() interface{}
}

type Session interface {
	Key() *string
	UserID() *uint64
	User() User
	SetUser(User)
	Broadcast(name string, val interface{})
	LastRequest() time.Time
	UpdateLastRequest()
	SessionStore() SessionStore
	GetDatabaseConnection() (*sql.DB, error)

	AddFlash(severity, format string, parameters ...interface{})
	ResetFlash()
	DisplayFlash() []FlashMessage
	Flash() []FlashMessage //alias!
}

type Request interface {
	Cleanup()
	Session() Session
	IsLoggedIn() bool
	ResetSession() error
	Method() string
	Redirect(to string)
	DB() (*sql.DB, error)
	GetContext() Context
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

type Response interface {
	WriteTo(w io.Writer) error
	ContentType() string
	HTTPExtra(http.ResponseWriter)
}
