package actions

import (
	"fmt"
	"github.com/daemonl/go_gsd/shared"
)

type CustomQuery struct {
	Core Core
}

type customQueryRequest struct {
	QueryName  string        `json:"queryName"`
	Parameters []interface{} `json:"parameters"`
}

func (q *CustomQuery) RequestDataPlaceholder() interface{} {
	r := customQueryRequest{}
	return &r
}

func (r *CustomQuery) Handle(request Request, requestData interface{}) (shared.IResponse, error) {
	cqr, ok := requestData.(*customQueryRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}
	model := r.Core.GetModel()

	customQuery, ok := model.CustomQueries[cqr.QueryName]
	if !ok {
		return nil, fmt.Errorf("No custom query called '%s'", cqr.QueryName)
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	results, err := customQuery.Run(db, cqr.Parameters)
	if err != nil {
		return nil, err
	}

	return JSON(results), nil
}
