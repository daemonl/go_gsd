package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
)

type SelectQuery struct {
	Core *GSDCore
}

func (q *SelectQuery) GetRequestObject() interface{} {
	r := databath.RawQueryConditions{}
	return &r
}

func (r *SelectQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	fmt.Println("HR SELECT")
	rawQueryCondition, ok := requestObject.(*databath.RawQueryConditions)
	if !ok {
		fmt.Println("Not Correct Type")
		return
	}
	queryConditions, err := rawQueryCondition.TranslateToQuery()
	if err != nil {
		fmt.Println("E", err)
		return
	}
	_ = queryConditions

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["me"] = os.Session.User.Id

	query, err := databath.GetQuery(&context, r.Core.Model, queryConditions)
	if err != nil {
		fmt.Println("E", err)
		return
	}
	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		fmt.Println("E", err)
		return
	}

	allRows, err := query.RunQueryWithResults(r.Core.Bath, sqlString, parameters)
	if err != nil {
		fmt.Println("E", err)
		return
	}
	fmt.Println("SEND")
	os.SendObject("result", responseId, allRows)

}
