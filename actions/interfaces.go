package actions

import (
	"database/sql"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_gsd/torch"
)

type Request interface {
	GetSession() *torch.Session
	Broadcast(functionName string, object interface{})
	GetContext() databath.Context
	DB() (*sql.DB, error)
}

type Handler interface {
	RequestDataPlaceholder() interface{}
	HandleRequest(req Request, requestData interface{}) (interface{}, error)
}

type Core interface {
	DoHooksPreAction(db *sql.DB, as *shared_structs.ActionSummary)
	DoHooksPostAction(db *sql.DB, as *shared_structs.ActionSummary)
	GetModel() *databath.Model
	RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
	SendMail(to string, subject string, body string)
}
