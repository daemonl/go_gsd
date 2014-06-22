package actions

import ()

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

func (r *DynamicHandler) HandleRequest(request Request, requestData interface{}) (interface{}, error) {

	cqr, ok := requestData.(*dynamicRequest)
	if !ok {
		return nil, ErrF("Request Type Mismatch")
	}

	db, err := request.DB()
	if err != nil {
		return nil, err
	}

	model := r.Core.GetModel()

	//dr := r.Runner

	fnConfig, ok := model.DynamicFunctions[cqr.FunctionName]
	if !ok {
		return nil, ErrF("No registered dynamic function named '%s'", cqr.FunctionName)
	}

	return r.Core.RunDynamic(fnConfig.Filename, cqr.Parameters, db)
}
