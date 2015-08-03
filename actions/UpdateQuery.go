package actions

import (
	"fmt"
	"log"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/components"
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
	var hookContext *components.HookContext
	if updateRequest.Conditions.Pk != nil { // Not for Bulk requests
		hookContext = &components.HookContext{
			DB: db,
			ActionSummary: &shared.ActionSummary{
				UserId:     *session.UserID(),
				Action:     "update",
				Collection: *updateRequest.Conditions.Collection,
				Pk:         *updateRequest.Conditions.Pk,
				Fields:     updateRequest.Changeset,
			},
			Session: request.Session(),
			Core:    r.Core,
		}
		r.Core.DoPreHooks(hookContext)
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

	if hookContext != nil { // Not for Bulk Requests, as above
		go r.Core.DoPostHooks(hookContext)
	}
	return JSON(map[string]int64{"affected": rows}), nil
}
