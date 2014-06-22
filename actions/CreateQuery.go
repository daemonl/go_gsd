package actions

import (
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared_structs"
)

type CreateQuery struct {
	Core Core
}

type createRequest struct {
	Values     map[string]interface{}
	Collection string
}

type createResult struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	InsertId int64  `json:"id"`
}

func (q *CreateQuery) RequestDataPlaceholder() interface{} {
	r := createRequest{}
	return &r
}

func (r *CreateQuery) HandleRequest(request Request, requestObject interface{}) (interface{}, error) {
	createRequest, ok := requestObject.(*createRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}

	session := request.GetSession()

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	qc := databath.GetMinimalQueryConditions(createRequest.Collection, "form")

	model := r.Core.GetModel()

	query, err := databath.GetQuery(request.GetContext(), model, qc, true)
	if err != nil {
		return nil, err
	}

	actionSummary := &shared_structs.ActionSummary{
		UserId:     session.User.Id,
		Action:     "create",
		Collection: createRequest.Collection,
		Pk:         0,
		Fields:     createRequest.Values,
	}

	r.Core.DoHooksPreAction(db, actionSummary)

	sqlString, parameters, err := query.BuildInsert(createRequest.Values)
	if err != nil {
		return nil, err
	}

	res, err := db.Exec(sqlString, parameters...)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	actionSummary.Pk = uint64(id)
	result := createResult{
		Status:   "OK",
		Message:  "Success",
		InsertId: id,
	}

	createObject := map[string]interface{}{
		"collection": createRequest.Collection,
		"id":         id,
		"object":     createRequest.Values,
	}

	go r.Core.DoHooksPostAction(db, actionSummary)
	go request.Broadcast("create", createObject)
	return result, nil
}
