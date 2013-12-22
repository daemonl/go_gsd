package view

import (
	"github.com/daemonl/go_gsd/torch"
	"html/template"
	"io"
	"log"
	"path/filepath"
)

type ViewManager struct {
	rootTemplate *template.Template
	pattern      string
	IncludeRoot  string
}

func GetViewManager(dir string, IncludeRoot string) *ViewManager {
	log.Println("Load View Manager")
	pattern := filepath.Join(dir, "*.html")
	viewManager := ViewManager{
		pattern:     pattern,
		IncludeRoot: IncludeRoot,
	}
	err := viewManager.Reload()
	if err != nil {
		panic(err)
	}
	return &viewManager
}

func (vm *ViewManager) Reload() error {
	templates, err := template.ParseGlob(vm.pattern)
	if err != nil {
		log.Println(err)
		return err
	}
	vm.rootTemplate = templates
	return nil
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
	Root    string
}

func (vh *ViewHandler) Handle(r *torch.Request) {
	d := ViewData{
		Session: r.Session,
		Data:    vh.Data,
		Root:    vh.Manager.IncludeRoot,
	}
	err := vh.Manager.Render(r.GetWriter(), vh.TemplateName, &d)
	if err != nil {
		log.Println(err.Error())
	}
	r.Session.ResetFlash()
}
