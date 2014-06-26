package actions

import (
	"fmt"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
)

type SelectQuery struct {
	Core Core
}

func (q *SelectQuery) RequestDataPlaceholder() interface{} {
	r := databath.RawQueryConditions{}
	return &r
}

func (r *SelectQuery) Handle(request Request, requestData interface{}) (shared.IResponse, error) {

	rawQueryCondition, ok := requestData.(*databath.RawQueryConditions)
	if !ok {
		return nil, fmt.Errorf("Request type mismatch")
	}
	queryConditions, err := rawQueryCondition.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	model := r.Core.GetModel()

	query, err := databath.GetQuery(request.GetContext(), model, queryConditions, false)
	if err != nil {
		return nil, err
	}

	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	allRows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}

	return JSON(allRows), nil

}
