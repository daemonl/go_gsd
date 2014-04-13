package actions

import (
	"github.com/daemonl/go_lib/databath"
)

type SelectQuery struct {
	Core Core
}

func (q *SelectQuery) GetRequestObject() interface{} {
	r := databath.RawQueryConditions{}
	return &r
}

func (r *SelectQuery) HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error) {

	rawQueryCondition, ok := requestObject.(*databath.RawQueryConditions)
	if !ok {
		return nil, ErrF("Request type mismatch")
	}
	queryConditions, err := rawQueryCondition.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	model := r.Core.GetModel()

	query, err := databath.GetQuery(ac.GetContext(), model, queryConditions, false)
	if err != nil {
		return nil, err
	}

	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	c := r.Core.GetConnection()
	db := c.GetDB()
	defer c.Release()

	allRows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}

	return allRows, nil

}
