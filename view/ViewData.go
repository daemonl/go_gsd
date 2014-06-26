package view

import (
	"github.com/daemonl/go_gsd/shared"
	"io"
	"log"
	"net/http"
)

type ViewData struct {
	Session      shared.ISession
	Data         interface{}
	D            map[string]interface{}
	Root         string
	TemplateName string
	Manager      *ViewManager
}

func (vd *ViewData) ContentType() string {
	return "text/html"
}

func (vd *ViewData) WriteTo(w io.Writer) error {
	log.Println("TEMPLATE NAME " + vd.TemplateName)
	err := vd.Manager.Render(w, vd.TemplateName, vd)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (vd *ViewData) HTTPExtra(w http.ResponseWriter) {}
