package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_lib/databath"
)

type CreateQuery struct {
	Model *databath.Model
	Bath  *databath.Bath
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

	query, err := databath.GetQuery(&context, r.Model, qc)
	if err != nil {
		fmt.Println(err)
		return
	}
	sqlString, err := query.BuildInsert(createRequest.Values)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := r.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()
	res, err := db.Exec(sqlString)
	if err != nil {
		fmt.Println(err)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		fmt.Println(err)
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
	go os.SendObjectToAll("create", createObject)
	os.SendObject("result", responseId, result)
}
