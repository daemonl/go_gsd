package reporter

import (
	"database/sql"
	"fmt"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/view"
)

type Reporter struct {
	ViewManager *view.TemplateManager
	Runner      *dynamic.DynamicRunner
	Model       *databath.Model
	Reports     map[string]ReportConfig
}

func (r *Reporter) Handle(request shared.IPathRequest) (shared.IResponse, error) {

	var name string
	var pk uint64
	var fname string
	err := request.ScanPath(&name, &pk, &fname)
	if err != nil {
		return nil, err
	}
	return r.GetReportHTMLWriter(name, pk, request.Session())
}

func (r *Reporter) GetReportHTMLWriter(name string, pk uint64, session shared.ISession) (shared.IResponse, error) {
	reportConfig, ok := r.Reports[name]
	if !ok {
		return nil, &ReportNotFoundError{fmt.Sprintf("Report %s not found", name)}
	}

	report := &Report{
		Session: session,
		Config:  &reportConfig,
		RootID:  pk,
		Core:    r,
	}

	writer, err := report.PrepareWriter()
	if err != nil {
		return nil, fmt.Errorf("Error preparing report %s[%d]: %s", name, pk, err.Error())
	}

	return writer, nil
}

func (r *Reporter) doSelect(db *sql.DB, rawQueryConditions *databath.RawQueryConditions, context *databath.MapContext) ([]map[string]interface{}, error) {
	queryConditions, err := rawQueryConditions.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	query, err := databath.GetQuery(context, r.Model, queryConditions, false)
	if err != nil {
		return nil, err
	}
	sqlString, _, parameters, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	allRows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}
	return allRows, nil
}
