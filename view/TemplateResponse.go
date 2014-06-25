package view

import (
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/torch"

	"fmt"
)

type Report struct {
	Session torch.Session
	Config  *TemplateConfig
	RootID  uint64
	Core    *TemplateWriter
}

func (report *Report) PrepareData() (*ViewData, error) {

	emailParameters := map[string]interface{}{}

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["id"] = report.RootID

	fieldset := "email"

	rawQueryCondition := databath.RawQueryConditions{
		Collection: &report.Config.Collection,
		Pk:         &report.RootID,
		Fieldset:   &fieldset,
	}

	db := report.Core.DB

	results, err := report.Core.DoSelect(db, &rawQueryCondition, &context)
	if err != nil {
		return nil, err
	}

	if len(results) < 1 {
		return nil, fmt.Errorf("No results found for core object")
	}

	emailParameters[report.Config.Collection] = results[0]
	for k, v := range results[0] {
		context.Fields["main."+k] = v
	}

	for key, qc := range report.Config.Queries {
		results2, err := report.Core.DoSelect(db, qc, &context)
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
		javascriptData, err = report.Core.Runner.Run(report.Config.ScriptName, scriptParameters, db)

		if err != nil {
			return nil, err
		}
	}

	data := &ViewData{
		Data:    emailParameters,
		D:       javascriptData,
		Root:    report.Core.ViewManager.IncludeRoot,
		Session: report.Session,
	}

	return data, nil
}
