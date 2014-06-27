package shared

type IUser interface {
	GetContext() IContext
	CheckPassword(string) (bool, error)
	ID() uint64
	Access() uint64
	WhoAmIObject() interface{}
	Username() string
}
