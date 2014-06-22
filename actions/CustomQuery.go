package actions

import ()

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

func (r *CustomQuery) HandleRequest(request Request, requestData interface{}) (interface{}, error) {
	cqr, ok := requestData.(*customQueryRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}
	model := r.Core.GetModel()

	customQuery, ok := model.CustomQueries[cqr.QueryName]
	if !ok {
		return nil, ErrF("No custom query called '%s'", cqr.QueryName)
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	results, err := customQuery.Run(db, cqr.Parameters)
	if err != nil {
		return nil, err
	}

	return results, nil
}
