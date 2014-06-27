package shared

import (
	"database/sql"
)

type ISessionStore interface {
	GetSession(key string) (ISession, error)
	NewSession() (ISession, error)
	DumpSessions()
	SetBroadcast(func(string, interface{}))
	Broadcast(string, interface{})
	GetDatabaseConnectionForSession(ISession) (*sql.DB, error)
}
