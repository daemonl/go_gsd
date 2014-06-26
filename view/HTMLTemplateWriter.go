package view

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_sweetpl"
)

type HTMLTemplateWriter struct {
	Session      shared.ISession
	Data         interface{}
	D            map[string]interface{}
	Root         string
	TemplateName string
	Sweetpl      *sweetpl.SweeTpl
}

func (vd *HTMLTemplateWriter) ContentType() string {
	return "text/html"
}

func (vd *HTMLTemplateWriter) WriteTo(w io.Writer) error {
	log.Println("TEMPLATE NAME " + vd.TemplateName)

	err := vd.Sweetpl.Render(w, vd.TemplateName, vd)
	if err != nil {
		return fmt.Errorf("Error rendering template: %s", err.Error())
	}
	return nil
}

func (vd *HTMLTemplateWriter) HTTPExtra(w http.ResponseWriter) {}
