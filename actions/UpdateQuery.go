package actions

import (
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_lib/databath"
)

type UpdateQuery struct {
	Core Core
}

type updateRequest struct {
	Conditions databath.RawQueryConditions `json:"query"`
	Changeset  map[string]interface{}      `json:"changeset"`
}

func (q *UpdateQuery) GetRequestObject() interface{} {
	r := updateRequest{}
	return &r
}

func (r *UpdateQuery) HandleRequest(ac ActionCore, requestObject interface{}) (interface{}, error) {
	updateRequest, ok := requestObject.(*updateRequest)
	if !ok {
		return nil, ErrF("Request type mismatch")
	}
	queryConditions, err := updateRequest.Conditions.TranslateToQuery()
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
	if updateRequest.Conditions.Pk != nil {
		actionSummary := &shared_structs.ActionSummary{
			UserId:     session.User.Id,
			Action:     "update",
			Collection: *updateRequest.Conditions.Collection,
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		r.Core.DoHooksPreAction(actionSummary)
	}
	sqlString, parameters, err := query.BuildUpdate(updateRequest.Changeset)
	if err != nil {
		return nil, err
	}

	c := r.Core.GetConnection()
	db := c.GetDB()
	defer c.Release()
	resp, err := db.Exec(sqlString, parameters...)
	if err != nil {
		return nil, err
	}
	rows, _ := resp.RowsAffected()

	updateObject := map[string]interface{}{
		"collection": updateRequest.Conditions.Collection,
		"id":         updateRequest.Conditions.Pk,
	}

	go ac.Broadcast("update", updateObject)
	if updateRequest.Conditions.Pk != nil {
		actionSummary := &shared_structs.ActionSummary{
			UserId:     session.User.Id,
			Action:     "update",
			Collection: *updateRequest.Conditions.Collection,
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		go r.Core.DoHooksPostAction(actionSummary)

	}
	return map[string]int64{"affected": rows}, nil
}
