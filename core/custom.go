package core

import (
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
	"log"
)

type CustomQuery struct {
	Model *databath.Model
	Bath  *databath.Bath
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

	customQuery, ok := r.Model.CustomQueries[cqr.QueryName]
	if !ok {
		log.Printf("No query called '%s'", cqr.QueryName)
		return
	}

	results, err := customQuery.Run(r.Bath, cqr.Parameters)
	if err != nil {
		log.Println(err.Error())
		return
	}

	os.SendObject("result", responseId, results)

}
