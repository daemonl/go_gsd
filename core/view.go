package core

import (
	"github.com/daemonl/go_gsd/torch"
	"html/template"
	"io"
	"log"
	"path/filepath"
)

type ViewManager struct {
	rootTemplate *template.Template
}

func GetViewManager(dir string) *ViewManager {
	log.Println("Load View Manager")
	pattern := filepath.Join(dir, "*.html")
	templates := template.Must(template.ParseGlob(pattern))
	viewManager := ViewManager{
		rootTemplate: templates,
	}
	return &viewManager
}

func (vm *ViewManager) Render(w io.Writer, name string, data *ViewData) error {
	err := vm.rootTemplate.ExecuteTemplate(w, name, data)
	return err
}

type ViewHandler struct {
	Manager      *ViewManager
	TemplateName string
	Data         interface{}
}

type ViewData struct {
	Session *torch.Session
	Data    interface{}
}

func (vh *ViewHandler) Handle(r *torch.Request) {
	d := ViewData{
		Session: r.Session,
		Data:    vh.Data,
	}
	err := vh.Manager.Render(r.GetWriter(), vh.TemplateName, &d)
	if err != nil {
		log.Println(err.Error())
	}
	r.Session.ResetFlash()
}
