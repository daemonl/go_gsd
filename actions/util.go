package actions

import (
	"encoding/json"
	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/torch"
	"io"
)

type JSONResponse struct {
	obj interface{}
}

func JSON(obj interface{}) *JSONResponse {
	return &JSONResponse{
		obj: obj,
	}
}

func (j *JSONResponse) WriteTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(j.obj)
	return err
}

func (j *JSONResponse) ContentType() string {
	return "application/json"
}

type wrappedHandler struct {
	handler Handler
}

func AsRouterHandler(h Handler) router.Handler {
	return &wrappedHandler{
		handler: h,
	}
}

func (wh *wrappedHandler) Handle(request torch.Request) (router.Response, error) {

	requestData := wh.handler.RequestDataPlaceholder()
	_, r := request.GetRaw()

	// Decode request body JSON to object
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(requestData)
	if err != nil {
		return nil, err
	}

	// Do the request
	return wh.handler.Handle(request, requestData)

}
