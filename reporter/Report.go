package reporter

import (
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/view"
)

type Report struct {
	Session shared.ISession
	RootID  uint64
	Config  *ReportConfig
	Core    *Reporter
}

type ReportConfig struct {
	TemplateFile string                                  `json:"templateFile"`
	Collection   string                                  `json:"collection"`
	Queries      map[string]*databath.RawQueryConditions `json:"queries"`
	ScriptName   string                                  `json:"script"`
}

type ReportNotFoundError struct {
	Message string
}

func (e *ReportNotFoundError) Error() string {
	return e.Message
}
func (e *ReportNotFoundError) GetUserDescription() string {
	return e.Message
}
func (e *ReportNotFoundError) GetHTTPStatus() int {
	return 404
}

func (report *Report) PrepareWriter() (*view.HTMLTemplateWriter, error) {

	emailParameters := map[string]interface{}{}

	fieldset := "email"
	db, err := report.Session.GetDatabaseConnection()
	if err != nil {
		return nil, err
	}

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	if len(report.Config.Collection) > 0 {

		context.Fields["id"] = report.RootID

		rawQueryCondition := databath.RawQueryConditions{
			Collection: &report.Config.Collection,
			Pk:         &report.RootID,
			Fieldset:   &fieldset,
		}

		results, err := report.Core.doSelect(db, &rawQueryCondition, &context)
		if err != nil {
			return nil, err
		}

		if len(results) < 1 {
			return nil, &ReportNotFoundError{"No results found for core object"}
		}

		emailParameters[report.Config.Collection] = results[0]

		for k, v := range results[0] {
			context.Fields["main."+k] = v
		}

	}

	for key, qc := range report.Config.Queries {
		results2, err := report.Core.doSelect(db, qc, &context)
		if err != nil {
			return nil, err
		}
		emailParameters[key] = results2
	}

	javascriptData := map[string]interface{}{}

	if len(report.Config.ScriptName) > 0 {
		queryMap := map[string]string{}
		//_, req := request.GetRaw()
		//query := req.URL.Query().Get(key)
		scriptParameters := map[string]interface{}{
			"context":      context.Fields,
			"id":           report.RootID,
			"fieldset":     fieldset,
			"requestQuery": queryMap,
			"queries":      emailParameters,
		}
		javascriptData, err = report.Core.Runner.RunScript(report.Config.ScriptName, scriptParameters, db)

		if err != nil {
			return nil, err
		}
	}

	htmlWriter, err := report.Core.ViewManager.GetHTMLTemplateWriter(report.Config.TemplateFile, report.Session)
	if err != nil {
		return nil, err
	}

	htmlWriter.Data = emailParameters
	htmlWriter.D = javascriptData
	htmlWriter.Root = report.Core.ViewManager.IncludeRoot

	return htmlWriter, nil
}
