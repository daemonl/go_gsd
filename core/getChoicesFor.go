package core

import (
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
)

type ChoicesForQuery struct {
	Model *databath.Model
	Bath  *databath.Bath
}

type RawChoicesQuery struct {
	Collection string `json:"collection"`
	Pk         uint64 `json:"pk"`
	Field      string `json:"field"`
	Search     string `json:"search"`
}

func (q *ChoicesForQuery) GetRequestObject() interface{} {
	r := RawChoicesQuery{}
	return &r
}

func (r *ChoicesForQuery) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	cq, ok := requestObject.(*RawChoicesQuery)
	if !ok {
		return
	}

	collection, ok := r.Model.Collections[cq.Collection]
	if !ok {
		return
	}

	field, ok := collection.Fields[cq.Field]
	if !ok {
		return
	}

	refField, ok := field.(*databath.FieldRef)
	if !ok {
		return
	}

	searchCollectionName := refField.Collection
	str_identity := "identity"
	var int_limit int64 = 10
	search := map[string]string{
		"*": cq.Search,
	}
	searchQueryConditions := databath.RawQueryConditions{
		Collection: &searchCollectionName,
		Fieldset:   &str_identity,
		Limit:      &int_limit,
		Search:     search,
	}

	sq := SelectQuery{
		Model: r.Model,
		Bath:  r.Bath,
	}
	sq.HandleRequest(os, &searchQueryConditions, responseId)
}
