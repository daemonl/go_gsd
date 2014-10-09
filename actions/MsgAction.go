package actions

import (
	"fmt"

	"github.com/daemonl/go_gsd/shared"
)

type MsgAction struct {
	Core Core
}

type MsgActionRequest struct {
	Msg string `json: "msg"`
}

func (q *MsgAction) RequestDataPlaceholder() interface{} {
	r := MsgActionRequest{}
	return &r
}

func (r *MsgAction) Handle(request Request, requestData interface{}) (shared.IResponse, error) {
	cqr, ok := requestData.(*MsgActionRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}
	request.Broadcast("msg", cqr.Msg)
	return JSON(cqr.Msg), nil
}
