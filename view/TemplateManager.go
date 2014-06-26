package view

import (
	"html/template"
	"log"
	"strings"

	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_sweetpl"

	"github.com/russross/blackfriday"
)

type TemplateManager struct {
	Sweetpl      *sweetpl.SweeTpl
	TemplateRoot string
	IncludeRoot  string
}

func GetTemplateManager(dir string, IncludeRoot string) *TemplateManager {
	log.Println("Load View Manager. Template Root: " + IncludeRoot)
	//pattern := filepath.Join(dir, "*.html")
	templateManager := TemplateManager{
		TemplateRoot: dir,
		IncludeRoot:  IncludeRoot,
	}
	err := templateManager.load()
	if err != nil {
		panic(err)
	}
	return &templateManager
}

func (vm *TemplateManager) GetHTMLTemplateWriter(templateName string, session shared.ISession) (*HTMLTemplateWriter, error) {
	writer := &HTMLTemplateWriter{
		Session:      session,
		TemplateName: templateName,
		Sweetpl:      vm.Sweetpl,
	}
	return writer, nil
}

type simpleTemplateHandler struct {
	name    string
	manager *TemplateManager
}

func (th *simpleTemplateHandler) Handle(request shared.IRequest) (shared.IResponse, error) {
	writer, err := th.manager.GetHTMLTemplateWriter(th.name, request.Session())
	if err != nil {
		return nil, err
	}

	return writer, nil
}

func (vm *TemplateManager) GetSimpleTemplateHandler(name string) shared.IHandler {
	return &simpleTemplateHandler{
		name:    name,
		manager: vm,
	}
}

func (vm *TemplateManager) load() error {

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

/*
type ViewHandler struct {
	Manager      *ViewManager
	TemplateName string
	Data         interface{}
	JsData       map[string]interface{}
}

func (vh *ViewHandler) Handle(r shared.IRequest) (shared.IResponse, error) {
	session := r.Session()
	if session == nil {
		panic("NILL SESSION")
	}
	d := ViewData{
		Session:      session,
		Data:         vh.Data,
		D:            vh.JsData,
		Root:         vh.Manager.IncludeRoot,
		TemplateName: vh.TemplateName,
		Manager:      vh.Manager,
	}

	return &d, nil
}
*/
