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

	session := ac.GetSession()
	model := r.Core.GetModel()

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["me"] = session.User.Id

	query, err := databath.GetQuery(&context, model, queryConditions)
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
