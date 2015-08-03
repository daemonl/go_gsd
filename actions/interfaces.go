package actions

import (
	"database/sql"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/components"
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
	components.Core
	GetModel() *databath.Model
}
