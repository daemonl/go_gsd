package torch

type ActionCore interface {
	GetSession() *Session
	Broadcast(functionName string, object interface{})
}

type Handler interface {
	GetRequestObject() interface{}
	HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error)
}
