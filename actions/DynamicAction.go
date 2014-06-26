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

/*
func GetDynamicHandlerFromCore(core *GSDCore) *DynamicHandler {

	runner := &dynamic.DynamicRunner{
		DataBath:      core.Bath,
		BaseDirectory: core.Config.ScriptDirectory, // "/home/daemonl/schkit/impl/pov/script/",
		SendMail:      core.SendMail,
	}

	return &DynamicHandler{
		Core:   core,
		Runner: runner,
	}
}
*/

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

	resp, err := r.Core.RunDynamic(fnConfig.Filename, cqr.Parameters, db)
	if err != nil {
		return nil, err
	}
	return JSON(resp), nil
}
