package torch

import (
	"github.com/daemonl/databath"
	"time"
)

type ActionCore interface {
	GetSession() *Session
	Broadcast(functionName string, object interface{})
}

type Handler interface {
	GetRequestObject() interface{}
	HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error)
}

type SessionStore interface {
	GetSession(key string) (Session, error)
	NewSession() (Session, error)
}

type User interface {
	GetContext() databath.Context
	CheckPassword(string) (bool, error)
	ID() uint64
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

	AddFlash(severity, format string, parameters ...interface{})
	ResetFlash()
	DisplayFlash() []FlashMessage
}

type Request interface {
	Cleanup()
	Session() Session
	IsLoggedIn() bool
	ResetSession() error
	Method() string
	Redirect(to string)

	PostValueString(name string) string
}
