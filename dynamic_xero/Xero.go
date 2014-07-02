package dynamic_xero

import (
	"encoding/json"
	"github.com/daemonl/go_xero"
	"github.com/daemonl/go_xero/xero_objects"
)

type DynamicXero struct {
	Xero *xero.Xero
}

func (dx *DynamicXero) PostInvoice(raw string) (string, error) {
	invoice := &xero_objects.Invoice{}
	err := json.Unmarshal([]byte(raw), invoice)
	if err != nil {
		return "", err
	}
	res, err := dx.Xero.PostInvoice(invoice)
	if err != nil {
		return "", err
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}
