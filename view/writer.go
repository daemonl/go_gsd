package view

import (
	"database/sql"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/dynamic"
)

type TemplateConfig struct {
	TemplateFile string                                  `json:"templateFile"`
	Collection   string                                  `json:"collection"`
	Queries      map[string]*databath.RawQueryConditions `json:"queries"`
	ScriptName   string                                  `json:"script"`
}

type TemplateWriter struct {
	Model       *databath.Model
	ViewManager *ViewManager
	Runner      *dynamic.DynamicRunner
	DB          *sql.DB
}

func (h *TemplateWriter) DoSelect(db *sql.DB, rawQueryConditions *databath.RawQueryConditions, context *databath.MapContext) ([]map[string]interface{}, error) {
	queryConditions, err := rawQueryConditions.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	query, err := databath.GetQuery(context, h.Model, queryConditions, false)
	if err != nil {
		return nil, err
	}
	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	allRows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}
	return allRows, nil

}
