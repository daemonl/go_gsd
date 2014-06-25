package view

import (
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_sweetpl"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"log"

	"strings"
)

type ViewManager struct {
	Sweetpl      *sweetpl.SweeTpl
	TemplateRoot string
	IncludeRoot  string
}

func GetViewManager(dir string, IncludeRoot string) *ViewManager {
	log.Println("Load View Manager. Template Root: " + IncludeRoot)
	//pattern := filepath.Join(dir, "*.html")
	viewManager := ViewManager{
		TemplateRoot: dir,
		IncludeRoot:  IncludeRoot,
	}
	err := viewManager.Reload()
	if err != nil {
		panic(err)
	}
	return &viewManager
}

func (vm *ViewManager) Reload() error {

	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	//htmlFlags |= blackfriday.HTML_SKIP_SCRIPT
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK

	tpl := &sweetpl.SweeTpl{
		Loader: &sweetpl.DirLoader{
			BasePath: vm.TemplateRoot,
		},
		FuncMap: template.FuncMap{
			"markdown": func(in string) template.HTML {
				output := blackfriday.Markdown([]byte(in), renderer, extensions)
				return template.HTML(string(output))
			},
			"htmlLineBreaks": func(in ...interface{}) template.HTML {
				if len(in) < 1 {
					return template.HTML("")
				}
				inStr, ok := in[0].(string)
				if !ok {
					return ""
				}
				safe := template.HTMLEscapeString(inStr)
				return template.HTML(strings.Replace(safe, "\n", "<br/>", -1))
			},
		},
	}

	vm.Sweetpl = tpl
	return nil
}

func (vm *ViewManager) Render(w io.Writer, name string, data *ViewData) error {
	log.Printf("Render %s\n", name)
	err := vm.Sweetpl.Render(w, name, data)
	return err
}

type ViewHandler struct {
	Manager      *ViewManager
	TemplateName string
	Data         interface{}
	JsData       map[string]interface{}
}

type ViewData struct {
	Session torch.Session
	Data    interface{}
	D       map[string]interface{}
	Root    string
}

func (vh *ViewHandler) Handle(r torch.Request) {
	session := r.Session()
	if session == nil {
		panic("NILL SESSION")
	}
	d := ViewData{
		Session: session,
		Data:    vh.Data,
		D:       vh.JsData,
		Root:    vh.Manager.IncludeRoot,
	}
	w, _ := r.GetRaw()
	err := vh.Manager.Render(w, vh.TemplateName, &d)
	if err != nil {
		log.Println(err.Error())
	}
	//r.Session.ResetFlash()
}
