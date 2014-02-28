package core

import (
	"github.com/daemonl/go_gsd/socket"

	"log"
)

type CustomQuery struct {
	Core *GSDCore
}

type customQueryRequest struct {
	QueryName  string        `json:"queryName"`
	Parameters []interface{} `json:"parameters"`
}

func (q *CustomQuery) GetRequestObject() interface{} {
	r := customQueryRequest{}
	return &r
}

func (r *CustomQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	cqr, ok := requestObject.(*customQueryRequest)
	if !ok {
		return
	}
	log.Println("CUSTOM")

	customQuery, ok := r.Core.Model.CustomQueries[cqr.QueryName]
	if !ok {
		log.Printf("No query called '%s'", cqr.QueryName)
		return
	}

	results, err := customQuery.Run(r.Core.Bath, cqr.Parameters)
	if err != nil {
		log.Println(err.Error())
		os.SendError(responseId, err)
		return
	}

	os.SendObject("result", responseId, results)

}
