package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
)

type SelectQuery struct {
	Model *databath.Model
	Bath  *databath.Bath
}

func (q *SelectQuery) GetRequestObject() interface{} {
	r := databath.RawQueryConditions{}
	return &r
}

func (r *SelectQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	rawQueryCondition, ok := requestObject.(*databath.RawQueryConditions)
	if !ok {
		return
	}
	queryConditions, err := rawQueryCondition.TranslateToQuery()
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = queryConditions
	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	query, err := databath.GetQuery(&context, r.Model, queryConditions)
	if err != nil {
		fmt.Println(err)
		return
	}
	sqlString, err := query.BuildSelect()
	if err != nil {
		fmt.Println(err)
		return
	}

	allRows, err := query.RunQueryWithResults(r.Bath, sqlString)
	if err != nil {
		fmt.Println(err)
		return
	}
	os.SendObject("result", responseId, allRows)

}
