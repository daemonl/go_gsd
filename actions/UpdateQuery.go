package actions

import (
	"fmt"
	"log"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
)

type UpdateQuery struct {
	Core Core
}

type updateRequest struct {
	Conditions databath.RawQueryConditions `json:"query"`
	Changeset  map[string]interface{}      `json:"changeset"`
}

func (q *UpdateQuery) RequestDataPlaceholder() interface{} {
	r := updateRequest{}
	return &r
}

func (r *UpdateQuery) Handle(request Request, requestData interface{}) (shared.IResponse, error) {
	updateRequest, ok := requestData.(*updateRequest)
	if !ok {
		return nil, fmt.Errorf("Request type mismatch")
	}
	queryConditions, err := updateRequest.Conditions.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	session := request.Session()
	model := r.Core.GetModel()
	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	query, err := databath.GetQuery(request.GetContext(), model, queryConditions, true)
	if err != nil {
		return nil, err
	}
	if updateRequest.Conditions.Pk != nil {
		actionSummary := &shared.ActionSummary{
			UserId:     *session.UserID(),
			Action:     "update",
			Collection: *updateRequest.Conditions.Collection,
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		r.Core.DoHooksPreAction(db, actionSummary, session)
	}
	sqlString, parameters, err := query.BuildUpdate(updateRequest.Changeset)
	if err != nil {
		return nil, err
	}
	log.Printf("Run: %s %v\n", sqlString, parameters)
	resp, err := db.Exec(sqlString, parameters...)
	if err != nil {
		return nil, err
	}
	rows, _ := resp.RowsAffected()

	updateObject := map[string]interface{}{
		"collection": updateRequest.Conditions.Collection,
		"id":         updateRequest.Conditions.Pk,
	}

	go request.Broadcast("update", updateObject)
	if updateRequest.Conditions.Pk != nil {
		actionSummary := &shared.ActionSummary{
			UserId:     *session.UserID(),
			Action:     "update",
			Collection: *updateRequest.Conditions.Collection,
			Pk:         *updateRequest.Conditions.Pk,
			Fields:     updateRequest.Changeset,
		}
		go r.Core.DoHooksPostAction(db, actionSummary, session)

	}
	return JSON(map[string]int64{"affected": rows}), nil
}
