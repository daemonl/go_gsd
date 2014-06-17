package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	"github.com/daemonl/go_lib/databath/types"
	"log"
	"net/url"
	"strings"
	"time"
)

type CSVHandler struct {
	Model *databath.Model
}

func GetCsvHandler( Model *databath.Model) *CSVHandler {
	fh := CSVHandler{
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
	var neg1 int64 = -1
	rawQuery.Limit = &neg1
	rawQuery.Offset = nil

	qc, err := rawQuery.TranslateToQuery()
	if err != nil {
		requestTorch.DoError(err)
		return
	}

	query, err := databath.GetQuery(requestTorch.GetContext(), h.Model, qc, false)
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

	db, err := requestTorch.DB()
	if err != nil {
		requestTorch.DoError(err)
		return
	}

	rows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
		return
	}

	filename := *rawQuery.Collection + "-" + time.Now().Format("2006-01-02") + ".csv"
	w.Header().Add("content-type", "text/csv")
	w.Header().Add("Content-Disposition", "attachment; filename="+filename)
	csvWriter := csv.NewWriter(w)

	if len(rows) < 0 {
		return
	}

	mappedFields, err := query.GetFields()

	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
	}

	allColNames, err := query.GetColNames()
	if err != nil {
		log.Print(err)
		requestTorch.DoError(err)
	}

	colNames := make([]string, 0, 0)
	for _, colName := range allColNames {
		if colName == "id" {
			continue
		}
		if strings.HasPrefix(colName, "-") {
			continue
		}
		if strings.HasPrefix(colName, "#") {
			continue
		}
		if strings.HasSuffix(colName, "id") {
			continue
		}

		colNames = append(colNames, colName)

	}

	csvWriter.Write(colNames)

	// TODO: Type assertions etc are done on each row... this seems rather inefficient.

	for _, row := range rows {

		record := make([]string, len(colNames), len(colNames))
		for i, colName := range colNames {
			v, ok := row[colName]

			if ok && v != nil {
				field, ok := mappedFields[colName]
				if !ok {
					log.Printf("No field %s in %#v\n", colName, mappedFields)
				}
				switch fieldImpl := field.Impl.(type) {
				case *types.FieldEnum:
					record[i] = fieldImpl.Choices[v.(string)]
				default:
					record[i] = fmt.Sprintf("%v", v)
				}
			} else {
				record[i] = ""
			}
		}

		csvWriter.Write(record)
	}
	csvWriter.Flush()
}
