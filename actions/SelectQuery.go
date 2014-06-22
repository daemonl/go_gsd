package actions

import (
	"github.com/daemonl/databath"
)

type SelectQuery struct {
	Core Core
}

func (q *SelectQuery) RequestDataPlaceholder() interface{} {
	r := databath.RawQueryConditions{}
	return &r
}

func (r *SelectQuery) HandleRequest(request Request, requestData interface{}) (interface{}, error) {

	rawQueryCondition, ok := requestData.(*databath.RawQueryConditions)
	if !ok {
		return nil, ErrF("Request type mismatch")
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
	if err != nil{
		return nil, err
	}

	allRows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}

	return allRows, nil

}
