package actions

import (
	"database/sql"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
)

type Request interface {
	Session() shared.ISession
	Broadcast(functionName string, object interface{})
	GetContext() shared.IContext
	//URLMatch(dest ...interface{}) error
	DB() (*sql.DB, error)
}

type Handler interface {
	RequestDataPlaceholder() interface{}
	Handle(req Request, requestData interface{}) (shared.IResponse, error)
}

type Core interface {
	DoHooksPreAction(db *sql.DB, as *shared.ActionSummary, session shared.ISession)
	DoHooksPostAction(db *sql.DB, as *shared.ActionSummary, session shared.ISession)
	GetModel() *databath.Model
	RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
	SendMail(to string, subject string, body string)
}
