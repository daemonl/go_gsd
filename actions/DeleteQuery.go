package actions

import (
	"github.com/daemonl/go_lib/databath"
	"strings"
)

type DeleteQuery struct {
	Core Core
}

type deleteRequest struct {
	Id         uint64 `json:"pk"`
	Collection string `json:"collection"`
}

func (q *DeleteQuery) GetRequestObject() interface{} {
	r := deleteRequest{}
	return &r
}

func (r *DeleteQuery) HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error) {
	deleteRequest, ok := requestObject.(*deleteRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}

	session := ac.GetSession()
	model := r.Core.GetModel()

	qc := databath.GetMinimalQueryConditions(deleteRequest.Collection, "form")

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["me"] = session.User.Id

	query, err := databath.GetQuery(&context, model, qc)
	if err != nil {
		return nil, err
	}

	c := r.Core.GetConnection()
	db := c.GetDB()
	defer c.Release()

	deleteCheckResult, err := query.CheckDelete(db, deleteRequest.Id)
	if err != nil {
		return nil, err
	}

	if deleteCheckResult.Prevents {
		return nil, ErrF("Could not delete, as owners exist: \n%s", strings.Join(deleteCheckResult.GetIssues(), "\n"))
	}

	err = deleteCheckResult.ExecuteRecursive(db)
	if err != nil {
		return nil, err
	}

	sqlString, err := query.BuildDelete(deleteRequest.Id)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(sqlString)
	if err != nil {
		return nil, err
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
	go ac.Broadcast("delete", deleteObject)

	return result, nil

}
