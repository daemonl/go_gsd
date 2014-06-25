package torch

import ()

type Context interface {
	GetValueFor(string) interface{}
	GetUserLevel() (isApplication bool, userAccessLevel uint64)
}

type mapContext struct {
	IsApplication   bool
	UserAccessLevel uint64
	Fields          map[string]interface{}
}

func (mc *mapContext) GetUserLevel() (bool, uint64) {
	return mc.IsApplication, mc.UserAccessLevel
}

func (mc *mapContext) GetValueFor(key string) interface{} {
	val, ok := mc.Fields[key]
	if !ok {
		return key
	}
	return val
}
