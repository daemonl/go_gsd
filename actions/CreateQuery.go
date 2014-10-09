package actions

import (
	"fmt"
	"log"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
)

type CreateQuery struct {
	Core Core
}

type createRequest struct {
	Values     map[string]interface{}
	Collection string
	Fieldset   string
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

func (r *CreateQuery) Handle(request Request, requestObject interface{}) (shared.IResponse, error) {
	createRequest, ok := requestObject.(*createRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}

	session := request.Session()

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	if len(createRequest.Fieldset) < 1 {
		createRequest.Fieldset = "form"
	}

	qc := databath.GetMinimalQueryConditions(createRequest.Collection, createRequest.Fieldset)

	model := r.Core.GetModel()

	query, err := databath.GetQuery(request.GetContext(), model, qc, true)
	if err != nil {
		return nil, err
	}

	actionSummary := &shared.ActionSummary{
		UserId:     *session.UserID(),
		Action:     "create",
		Collection: createRequest.Collection,
		Pk:         0,
		Fields:     createRequest.Values,
	}

	r.Core.DoHooksPreAction(db, actionSummary, request.Session())

	sqlString, parameters, err := query.BuildInsert(createRequest.Values)
	if err != nil {
		return nil, err
	}

	res, err := db.Exec(sqlString, parameters...)
	if err != nil {
		log.Printf("ERROR in Exec %s: %s\n", sqlString, err.Error())
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

	go r.Core.DoHooksPostAction(db, actionSummary, request.Session())
	go request.Broadcast("create", createObject)
	return JSON(result), nil
}
