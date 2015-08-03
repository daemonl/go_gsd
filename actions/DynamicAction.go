package actions

import (
	"fmt"

	"github.com/daemonl/go_gsd/shared"
)

type DynamicHandler struct {
	Core Core
	//Runner *dynamic.DynamicRunner
}
type dynamicRequest struct {
	FunctionName string                 `json:"function"`
	Parameters   map[string]interface{} `json:"parameters"`
}

func (q *DynamicHandler) RequestDataPlaceholder() interface{} {
	r := dynamicRequest{}
	return &r
}

func (r *DynamicHandler) Handle(request Request, requestData interface{}) (shared.IResponse, error) {

	cqr, ok := requestData.(*dynamicRequest)
	if !ok {
		return nil, fmt.Errorf("Request Type Mismatch")
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	model := r.Core.GetModel()

	//dr := r.Runner

	fnConfig, ok := model.DynamicFunctions[cqr.FunctionName]
	if !ok {
		return nil, fmt.Errorf("No registered dynamic function named '%s'", cqr.FunctionName)
	}

	if len(fnConfig.Access) > 0 {
		can := false
		access := request.Session().User().Access()
		for _, a := range fnConfig.Access {
			if a == access {
				can = true
				break
			}

		}
		if !can {
			return nil, fmt.Errorf("NOT AUTHORISED")
		}
	}

	resp, err := r.Core.RunScript(fnConfig.Filename, cqr.Parameters, db)
	if err != nil {
		return nil, err
	}
	return JSON(resp), nil
}
