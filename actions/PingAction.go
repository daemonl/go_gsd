package actions

import ()

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

func (r *PingAction) HandleRequest(request Request, requestData interface{}) (interface{}, error) {
	cqr, ok := requestData.(*PingActionRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}
	return cqr.Msg, nil
}
