package torch

import (
	"github.com/daemonl/databath"
)

type basicUser struct {
	IDinternal     uint64 `json:"id"`
	Username       string `json:"username"`
	password       string
	AccessInternal uint64 `json:"access"`
	SetOnNextLogin bool   `json:"set_on_next_login"`
}

func (u *basicUser) ID() uint64 {
	return u.IDinternal
}

func (u *basicUser) GetContext() databath.Context {

	context := &databath.MapContext{
		IsApplication:   false,
		UserAccessLevel: u.AccessInternal,
		Fields:          make(map[string]interface{}),
	}
	context.Fields["me"] = u.IDinternal
	context.Fields["user"] = u.IDinternal
	return context
}

func (u *basicUser) Access() uint64 {
	return u.AccessInternal
}
