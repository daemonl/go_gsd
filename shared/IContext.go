package shared

type IContext interface {
	GetValueFor(string) interface{}
	GetUserLevel() (isApplication bool, userAccessLevel uint64)
}
