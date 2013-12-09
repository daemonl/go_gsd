package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	"log"
	"net/url"
	"time"
)

type CSVHandler struct {
	Bath  *databath.Bath
	Model *databath.Model
}

func GetCsvHandler(location string, Bath *databath.Bath, Model *databath.Model) *CSVHandler {
	if Model == nil {
		panic("NO MODEL")
	}
	fh := CSVHandler{
		Bath:  Bath,
		Model: Model,
	}
	return &fh
}
func (h *CSVHandler) Handle(requestTorch *torch.Request) {

	var functionName string
	var queryStringQuery string

	err := requestTorch.UrlMatch(&functionName, &queryStringQuery)
	if err != nil {
		requestTorch.DoError(err)
	}
	w, _ := requestTorch.GetRaw()

	rawQuery := databath.RawQueryConditions{}

	jsonQuery, err := url.QueryUnescape(queryStringQuery)
	if err != nil {
		requestTorch.DoError(err)
		return
	}
	err = json.Unmarshal([]byte(jsonQuery), &rawQuery)
	if err != nil {
		requestTorch.DoError(err)
		return
	}

	qc, err := rawQuery.TranslateToQuery()
	if err != nil {
		requestTorch.DoError(err)
		return
	}

	context := databath.MapContext{
		Fields: make(map[string]interface{}),
	}
	query, err := databath.GetQuery(&context, h.Model, qc)
	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
		return
	}
	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
		return
	}

	rows, err := query.RunQueryWithResults(h.Bath, sqlString, parameters)
	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
		return
	}

	filename := *rawQuery.Collection + "-" + time.Now().Format("2006-01-02") + ".csv"
	w.Header().Add("content-type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename="+filename)
	csvWriter := csv.NewWriter(w)
	first := true
	for _, row := range rows {
		if first {
			header := make([]string, 0, 0)
			for k, _ := range row {
				header = append(header, fmt.Sprintf("%v", k))
			}
			csvWriter.Write(header)
			first = false
		}
		record := make([]string, 0, 0)
		for _, v := range row {
			record = append(record, fmt.Sprintf("%v", v))
		}
		csvWriter.Write(record)
	}
	csvWriter.Flush()

}
