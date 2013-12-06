package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
)

type updateRequest struct {
	Conditions databath.RawQueryConditions `json:"query"`
	Changeset  map[string]interface{}      `json:"changeset"`
}

type UpdateQuery struct {
	Model *databath.Model
	Bath  *databath.Bath
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
		return
	}
	_ = queryConditions
	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	query, err := databath.GetQuery(&context, r.Model, queryConditions)
	if err != nil {
		fmt.Printf("Error building query: %s\n", err.Error())
		return
	}
	sqlString, err := query.BuildUpdate(updateRequest.Changeset)
	if err != nil {
		fmt.Printf("Error executing update query: %s\n", err.Error())

		return
	}
	fmt.Println(sqlString)
	c := r.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()
	_, err = db.Query(sqlString)
	if err != nil {
		fmt.Println(err)
		return
	}

	updateObject := map[string]interface{}{
		"collection": updateRequest.Conditions.Collection,
		"id":         updateRequest.Conditions.Pk,
	}
	go os.SendObjectToAll("update", updateObject)
	if updateRequest.Conditions.Pk != nil {
		go r.Model.WriteHistory(r.Bath, os.Session.User.Id, "update", *updateRequest.Conditions.Collection, *updateRequest.Conditions.Pk)
	}
	os.SendObject("result", responseId, nil)
}
