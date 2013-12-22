package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_lib/databath"
)

type CreateQuery struct {
	Core *GSDCore
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

func (q *CreateQuery) GetRequestObject() interface{} {
	r := createRequest{}
	return &r
}

func (r *CreateQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	createRequest, ok := requestObject.(*createRequest)
	if !ok {
		return
	}

	qc := databath.GetMinimalQueryConditions(createRequest.Collection, "form")

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["me"] = os.Session.User.Id

	query, err := databath.GetQuery(&context, r.Core.Model, qc)
	if err != nil {
		fmt.Println(err)
		os.SendError(responseId, err)
		return
	}

	actionSummary := ActionSummary{
		User:       os.Session.User,
		Action:     "create",
		Collection: r.Core.Model.Collections[createRequest.Collection],
		Pk:         0,
		Fields:     createRequest.Values,
	}

	r.Core.Hooker.DoPreHooks(&actionSummary)

	sqlString, parameters, err := query.BuildInsert(createRequest.Values)
	if err != nil {
		fmt.Printf("Error building insert: %s", err)
		os.SendError(responseId, err)
		return
	}

	c := r.Core.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()
	fmt.Println(sqlString)
	res, err := db.Exec(sqlString, parameters...)
	if err != nil {
		fmt.Println(err)
		os.SendError(responseId, err)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		fmt.Println(err)
		os.SendError(responseId, err)
		return
	}
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
	//go doHooks(r.Core.Bath, r.Core.Model)

	go r.Core.Hooker.DoPostHooks(&actionSummary)
	go os.SendObjectToAll("create", createObject)
	os.SendObject("result", responseId, result)
}
