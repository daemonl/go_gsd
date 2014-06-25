package actions

import ()

type JSONResponse struct {
	obj interface{}
}

func JSON(obj interface{}) *JSONResponse {
	return &JSONResponse{
		obj: obj,
	}
}
