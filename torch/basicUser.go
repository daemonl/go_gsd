package torch

import (
	"database/sql"

	"github.com/daemonl/go_gsd/shared"
)

type basicUser struct {
	id             uint64
	username       string
	password       string
	access         uint64
	setOnNextLogin bool
}

func LoadBasicUser(rows *sql.Rows) (shared.IUser, error) {
	user := &basicUser{}
	err := rows.Scan(&user.id, &user.username, &user.password, &user.setOnNextLogin, &user.access)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *basicUser) ID() uint64 {
	return u.id
}

func (u *basicUser) GroupID() uint64 {
	return 0
}

func (u *basicUser) GetContext() shared.IContext {

	context := &mapContext{
		IsApplication:   false,
		UserAccessLevel: u.access,
		Fields:          make(map[string]interface{}),
	}
	context.Fields["me"] = u.id
	context.Fields["user"] = u.id
	return context
}

func (u *basicUser) Access() uint64 {
	return u.access
}

func (u *basicUser) Username() string {
	return u.username
}

func (u *basicUser) WhoAmIObject() interface{} {
	return map[string]interface{}{
		"id":             u.id,
		"username":       u.username,
		"access":         u.access,
		"setOnNextLogin": u.setOnNextLogin,
	}
}

func (u *basicUser) SetOnNextLogin() bool {
	return u.setOnNextLogin
}
