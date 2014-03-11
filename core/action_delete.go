package core

import (
	"errors"
	"fmt"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_lib/databath"
	"strings"
)

type DeleteQuery struct {
	Core *GSDCore
}

type deleteRequest struct {
	Id         uint64 `json:"pk"`
	Collection string `json:"collection"`
}

func (q *DeleteQuery) GetRequestObject() interface{} {
	r := deleteRequest{}
	return &r
}

func (r *DeleteQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	deleteRequest, ok := requestObject.(*deleteRequest)
	if !ok {
		return
	}

	qc := databath.GetMinimalQueryConditions(deleteRequest.Collection, "form")

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["me"] = os.Session.User.Id

	query, err := databath.GetQuery(&context, r.Core.Model, qc)
	if err != nil {
		os.SendError(responseId, err)
		return
	}

	c := r.Core.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()

	deleteCheckResult, err := query.CheckDelete(db, deleteRequest.Id)
	if err != nil {
		os.SendError(responseId, err)
		return
	}

	if deleteCheckResult.Prevents {
		msg := "Could not delete, as owners exist: \n" + strings.Join(deleteCheckResult.GetIssues(), "\n")
		os.SendError(responseId, errors.New(msg))
		return
	}

	err = deleteCheckResult.ExecuteRecursive(db)
	if err != nil {
		os.SendError(responseId, err)
		return
	}

	sqlString, err := query.BuildDelete(deleteRequest.Id)
	if err != nil {
		os.SendError(responseId, err)
		return
	}
	fmt.Println(sqlString)

	_, err = db.Exec(sqlString)
	if err != nil {
		os.SendError(responseId, err)
		return
	}

	result := createResult{
		Status:   "OK",
		Message:  "Success",
		InsertId: 0,
	}

	deleteObject := map[string]interface{}{
		"collection": deleteRequest.Collection,
		"id":         deleteRequest.Id,
	}
	go os.SendObjectToAll("delete", deleteObject)

	os.SendObject("result", responseId, result)

}
