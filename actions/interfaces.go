package actions

import (
	"database/sql"
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/databath"
)

type ActionCore interface {
	GetSession() *torch.Session
	Broadcast(functionName string, object interface{})
	GetContext() databath.Context
	DB() (*sql.DB, error)
}

type Handler interface {
	GetRequestObject() interface{}
	HandleRequest(os *ActionCore, requestObject interface{}) (interface{}, error)
}

type Core interface {
	DB(session *torch.Session) (*sql.DB, error)
	DoHooksPreAction(db *sql.DB, as *shared_structs.ActionSummary)
	DoHooksPostAction(db *sql.DB, as *shared_structs.ActionSummary)
	GetModel() *databath.Model
	RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
	SendMail(to string, subject string, body string)
}
