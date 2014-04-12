package actions

import (
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
)

type ActionCore interface {
	GetSession() *torch.Session
	Broadcast(functionName string, object interface{})
}

type Handler interface {
	GetRequestObject() interface{}
	HandleRequest(os *ActionCore, requestObject interface{}) (interface{}, error)
}

type Core interface {
	GetConnection() *databath.Connection
	DoHooksPreAction(as *shared_structs.ActionSummary)
	DoHooksPostAction(as *shared_structs.ActionSummary)
	GetModel() *databath.Model
	RunDynamic(filename string, parameters map[string]interface{}) (map[string]interface{}, error)
	SendMail(to string, subject string, body string)
}
