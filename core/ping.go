package core

import (
	"github.com/daemonl/go_gsd/socket"
)

type PingHandler struct{}

type ping struct {
}
type pong struct {
}

func (q *PingHandler) GetRequestObject() interface{} {
	r := ping{}
	return &r
}

func (r *PingHandler) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	result := pong{}
	os.SendObject("result", responseId, result)
}
