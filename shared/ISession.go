package shared

import (
	"database/sql"
	"time"
)

type ISession interface {
	Key() *string
	UserID() *uint64
	User() IUser
	SetUser(IUser)
	Broadcast(name string, val interface{})
	LastRequest() time.Time
	UpdateLastRequest()
	SessionStore() ISessionStore
	GetDatabaseConnection() (*sql.DB, error)
	ReleaseDB(*sql.DB)

	AddFlash(severity, format string, parameters ...interface{})
	ResetFlash()
	DisplayFlash() []FlashMessage
	Flash() []FlashMessage //alias!
}
