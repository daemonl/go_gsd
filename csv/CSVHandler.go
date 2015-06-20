package csv

import (
	"encoding/json"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"

	"log"
	"net/url"
	"strings"
	"time"
)

type CSVHandler struct {
	Model *databath.Model
}

func (h *CSVHandler) Handle(request shared.IPathRequest) (shared.IResponse, error) {

	var queryStringQuery string

	err := request.ScanPath(&queryStringQuery)
	if err != nil {
		return nil, err
	}

	rawQuery := databath.RawQueryConditions{}

	jsonQuery, err := url.QueryUnescape(queryStringQuery)
	if err != nil {
		return nil, err
	}
	log.Printf("Decode CSV Query: %s\n", jsonQuery)
	err = json.Unmarshal([]byte(jsonQuery), &rawQuery)
	if err != nil {
		return nil, err
	}
	var neg1 int64 = -1
	rawQuery.Limit = &neg1
	rawQuery.Offset = nil

	qc, err := rawQuery.TranslateToQuery()
	if err != nil {
		return nil, err
	}

	query, err := databath.GetQuery(request.GetContext(), h.Model, qc, false)
	if err != nil {
		return nil, err
	}
	sqlString, _, parameters, err := query.BuildSelect()
	if err != nil {
		return nil, err
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	rows, err := query.RunQueryWithResults(db, sqlString, parameters)
	if err != nil {
		return nil, err
	}

	filename := *rawQuery.Collection + "-" + time.Now().Format("2006-01-02") + ".csv"

	mappedFields, err := query.GetFields()

	if err != nil {
		log.Print(err)
		request.DoError(err)
	}

	allColNames, err := query.GetColNames()
	if err != nil {
		log.Print(err)
		request.DoError(err)
	}

	colNames := make([]string, 0, 0)
	for _, colName := range allColNames {
		if strings.HasPrefix(colName, "-") {
			continue
		}
		if strings.HasPrefix(colName, "#") {
			continue
		}

		colNames = append(colNames, colName)
	}

	resp := &csvResponse{
		rows:         rows,
		filename:     filename,
		colNames:     colNames,
		mappedFields: mappedFields,
	}

	return resp, nil

}
