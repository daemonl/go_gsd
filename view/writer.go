package view

import (
	"errors"
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	"io"
)

func errorf(format string, parameters ...interface{}) error {
	return errors.New(fmt.Sprintf(format, parameters...))
}

type TemplateConfig struct {
	TemplateFile string                                  `json:"templateFile"`
	Collection   string                                  `json:"collection"`
	Queries      map[string]*databath.RawQueryConditions `json:"queries"`
}

type TemplateWriter struct {
	Bath        *databath.Bath
	Model       *databath.Model
	ViewManager *ViewManager
}

func (h *TemplateWriter) DoSelect(rawQueryConditions *databath.RawQueryConditions) ([]map[string]interface{}, error) {
	queryConditions, err := rawQueryConditions.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}

	query, err := databath.GetQuery(&context, h.Model, queryConditions)
	if err != nil {
		return nil, err
	}
	sqlString, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	allRows, err := query.RunQueryWithResults(h.Bath, sqlString)
	if err != nil {
		return nil, err
	}
	return allRows, nil

}

func (h *TemplateWriter) Write(w io.Writer, requestTorch *torch.Request, templateConfig *TemplateConfig, rootId uint64) error {
	emailParameters := map[string]interface{}{}

	fieldset := "email"
	rawQueryCondition := databath.RawQueryConditions{
		Collection: &templateConfig.Collection,
		Pk:         &rootId,
		Fieldset:   &fieldset,
	}

	results, err := h.DoSelect(&rawQueryCondition)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if len(results) < 1 {
		return errorf("No results found for core object")
	}
	fmt.Println(results[0])
	emailParameters[templateConfig.Collection] = results[0]

	for key, qc := range templateConfig.Queries {
		results2, err := h.DoSelect(qc)
		if err != nil {
			fmt.Println(err)
			return err
		}
		emailParameters[key] = results2

	}

	data := ViewData{
		Session: requestTorch.Session,
		Data:    emailParameters,
	}

	err = h.ViewManager.Render(w, templateConfig.TemplateFile, &data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
