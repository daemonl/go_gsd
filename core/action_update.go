package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
)

type UpdateQuery struct {
	Core *GSDCore
}

type updateRequest struct {
	Conditions databath.RawQueryConditions `json:"query"`
	Changeset  map[string]interface{}      `json:"changeset"`
}

func (q *UpdateQuery) GetRequestObject() interface{} {
	r := updateRequest{}
	return &r
}

func (r *UpdateQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	updateRequest, ok := requestObject.(*updateRequest)
	if !ok {
		return
	}
	queryConditions, err := updateRequest.Conditions.TranslateToQuery()
	if err != nil {
		fmt.Printf("Error translating to query: %s\n", err.Error())
		os.SendError(responseId, err)
		return
	}
	_ = queryConditions
	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	query, err := databath.GetQuery(&context, r.Core.Model, queryConditions)
	if err != nil {
		fmt.Printf("Error building query: %s\n", err.Error())
		os.SendError(responseId, err)
		return
	}
	if updateRequest.Conditions.Pk != nil {
		actionSummary := ActionSummary{
			User:       os.Session.User,
			Action:     "update",
			Collection: r.Core.Model.Collections[*updateRequest.Conditions.Collection],
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		r.Core.Hooker.DoPreHooks(&actionSummary)
	}
	sqlString, parameters, err := query.BuildUpdate(updateRequest.Changeset)
	if err != nil {
		fmt.Printf("Error executing update query: %s\n", err.Error())
		os.SendError(responseId, err)
		return
	}
	fmt.Printf("SQL: %s, %#v\n", sqlString, parameters)
	c := r.Core.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()
	resp, err := db.Exec(sqlString, parameters...)
	if err != nil {
		os.SendError(responseId, err)
		fmt.Println(err)
		return
	}
	rows, _ := resp.RowsAffected()

	updateObject := map[string]interface{}{
		"collection": updateRequest.Conditions.Collection,
		"id":         updateRequest.Conditions.Pk,
	}
	go os.SendObjectToAll("update", updateObject)
	if updateRequest.Conditions.Pk != nil {
		actionSummary := ActionSummary{
			User:       os.Session.User,
			Action:     "update",
			Collection: r.Core.Model.Collections[*updateRequest.Conditions.Collection],
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		go r.Core.Hooker.DoPostHooks(&actionSummary)

	}
	os.SendObject("result", responseId, map[string]int64{"affected": rows})
}
