package shared

type IPathHandler interface {
	Handle(req IPathRequest) (IResponse, error)
}
