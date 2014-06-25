package actions

import (
	"fmt"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/router"
	"strings"
)

type DeleteQuery struct {
	Core Core
}

type deleteRequest struct {
	Id         uint64 `json:"pk"`
	Collection string `json:"collection"`
}

func (q *DeleteQuery) RequestDataPlaceholder() interface{} {
	r := deleteRequest{}
	return &r
}

func (r *DeleteQuery) Handle(request Request, requestData interface{}) (router.Response, error) {
	deleteRequest, ok := requestData.(*deleteRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}

	model := r.Core.GetModel()

	qc := databath.GetMinimalQueryConditions(deleteRequest.Collection, "form")

	query, err := databath.GetQuery(request.GetContext(), model, qc, true)
	if err != nil {
		return nil, err
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	deleteCheckResult, err := query.CheckDelete(db, deleteRequest.Id)
	if err != nil {
		return nil, err
	}

	if deleteCheckResult.Prevents {
		return nil, fmt.Errorf("Could not delete, as owners exist: \n%s", strings.Join(deleteCheckResult.GetIssues(), "\n"))
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
	go request.Broadcast("delete", deleteObject)

	return JSON(result), nil

}
