package view

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/databath"
	"io"
)

func errorf(format string, parameters ...interface{}) error {
	return errors.New(fmt.Sprintf(format, parameters...))
}

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

func (h *TemplateWriter) Write(w io.Writer, requestTorch *torch.Request, templateConfig *TemplateConfig, rootId uint64) error {
	emailParameters := map[string]interface{}{}

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	context.Fields["id"] = rootId

	fieldset := "email"
	rawQueryCondition := databath.RawQueryConditions{
		Collection: &templateConfig.Collection,
		Pk:         &rootId,
		Fieldset:   &fieldset,
	}

	db, err := requestTorch.DB()
	if err != nil {
		return err
	}

	results, err := h.DoSelect(db, &rawQueryCondition, &context)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if len(results) < 1 {
		return errorf("No results found for core object")
	}

	emailParameters[templateConfig.Collection] = results[0]
	for k, v := range results[0] {
		context.Fields["main."+k] = v
	}

	for key, qc := range templateConfig.Queries {
		results2, err := h.DoSelect(db, qc, &context)
		if err != nil {
			fmt.Println(err)
			return err
		}
		emailParameters[key] = results2
	}

	javascriptData := map[string]interface{}{}

	if len(templateConfig.ScriptName) > 0 {
		queryMap := map[string]string{}
		//_, req := requestTorch.GetRaw()
		//query := req.URL.Query().Get(key)
		scriptParameters := map[string]interface{}{
			"context":      context.Fields,
			"id":           rootId,
			"fieldset":     fieldset,
			"requestQuery": queryMap,
			"queries":      emailParameters,
		}
		javascriptData, err = h.Runner.Run(templateConfig.ScriptName, scriptParameters, db)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	data := ViewData{
		Data: emailParameters,
		D:    javascriptData,
		Root: h.ViewManager.IncludeRoot,
	}

	if requestTorch != nil {
		data.Session = requestTorch.Session
	}

	err = h.ViewManager.Render(w, templateConfig.TemplateFile, &data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
