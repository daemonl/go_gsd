package shared

type IHandler interface {
	Handle(req IRequest) (IResponse, error)
}
