package core

import (
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/socket"
	"log"
)

type DynamicHandler struct {
	Core   *GSDCore
	Runner *dynamic.DynamicRunner
}
type dynamicRequest struct {
	FunctionName string        `json:"function"`
	Parameters   []interface{} `json:"parameters"`
}

func GetDynamicHandlerFromCore(core *GSDCore) *DynamicHandler {

	runner := &dynamic.DynamicRunner{
		DataBath:      core.Bath,
		BaseDirectory: core.Config.ScriptDirectory, // "/home/daemonl/schkit/impl/pov/script/",
	}
	return &DynamicHandler{
		Core:   core,
		Runner: runner,
	}
}

func (q *DynamicHandler) GetRequestObject() interface{} {
	r := dynamicRequest{}
	return &r
}

func (r *DynamicHandler) HandleRequest(os *socket.OpenSocket, requestObject interface{}, responseId string) {
	dr := r.Runner
	cqr, ok := requestObject.(*dynamicRequest)
	if !ok {
		return
	}
	log.Println("DYNAMIC")

	fnConfig, ok := r.Core.Model.DynamicFunctions[cqr.FunctionName]
	if !ok {
		log.Printf("No registered dynamic function named '%s'", cqr.FunctionName)
		return
	}

	res, err := dr.Run(fnConfig.Filename)
	if err != nil {
		log.Println(err.Error())
		os.SendError(responseId, err)
		return
	}

	os.SendObject("result", responseId, res)
}
