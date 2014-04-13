package actions

import ()

type PingAction struct {
	Core Core
}

type PingActionRequest struct {
	Msg string `json: "msg"`
}

func (q *PingAction) GetRequestObject() interface{} {
	r := PingActionRequest{}
	return &r
}

func (r *PingAction) HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error) {
	cqr, ok := requestObject.(*PingActionRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}
	return cqr.Msg, nil
}
