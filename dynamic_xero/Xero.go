package dynamic_xero

import (
	"encoding/json"
	"github.com/daemonl/go_xero"
)

type DynamicXero struct {
	Xero *xero.Xero
}

func (dx *DynamicXero) Post(objectType string, obj interface{}, parameters ...string) (string, error) {
	res, err := dx.Xero.Post(objectType, obj, parameters...)
	if err != nil {
		return "", err
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}

