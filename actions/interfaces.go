package actions

import (
	"database/sql"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_gsd/torch"
)

type Request interface {
	Session() torch.Session
	Broadcast(functionName string, object interface{})
	GetContext() torch.Context
	//URLMatch(dest ...interface{}) error
	DB() (*sql.DB, error)
}

type Handler interface {
	RequestDataPlaceholder() interface{}
	Handle(req Request, requestData interface{}) (router.Response, error)
}

type Core interface {
	DoHooksPreAction(db *sql.DB, as *shared_structs.ActionSummary)
	DoHooksPostAction(db *sql.DB, as *shared_structs.ActionSummary)
	GetModel() *databath.Model
	RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
	SendMail(to string, subject string, body string)
}
