package core

import (
	"github.com/daemonl/go_gsd/socket"

	"github.com/daemonl/go_lib/databath"
	"github.com/daemonl/go_lib/databath/types"
)

type ChoicesForQuery struct {
	Core *GSDCore
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

	collection, ok := r.Core.Model.Collections[cq.Collection]
	if !ok {
		return
	}

	field, ok := collection.Fields[cq.Field]
	if !ok {
		return
	}

	refField, ok := field.Impl.(*types.FieldRef)
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
		Core: r.Core,
	}
	sq.HandleRequest(os, &searchQueryConditions, responseId)
}
