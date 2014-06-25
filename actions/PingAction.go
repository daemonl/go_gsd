package actions

import (
	"fmt"
	"github.com/daemonl/go_gsd/router"
)

type PingAction struct {
	Core Core
}

type PingActionRequest struct {
	Msg string `json: "msg"`
}

func (q *PingAction) RequestDataPlaceholder() interface{} {
	r := PingActionRequest{}
	return &r
}

func (r *PingAction) Handle(request Request, requestData interface{}) (router.Response, error) {
	cqr, ok := requestData.(*PingActionRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}
	return JSON(cqr.Msg), nil
}
