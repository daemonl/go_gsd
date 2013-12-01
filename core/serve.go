package core

import (
	"github.com/daemonl/go_gsd/pdf"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var parser *torch.Parser
var bath *databath.Bath

func checkHandle(requestTorch *torch.Request) {
	requestTorch.Writef("Session Key: %s\n\n", *requestTorch.Session.Key)
	if requestTorch.Session.User != nil {
		requestTorch.Writef("Username: %s", requestTorch.Session.User.Username)
	}
}

func signupHandle(requestTorch *torch.Request) {
	//username := requestTorch.PostValueString("username")
	password := requestTorch.PostValueString("password")
	hashStore := torch.HashPassword(password)
	requestTorch.Write(hashStore)
}

func pdfHandle(requestTorch *torch.Request) {
	in := strings.NewReader("<html><head></head><body style='font-family: ArialMT'>Hello World</body></html>")
	w := requestTorch.GetWriter()
	err := pdf.DoPdf("/Applications/wkhtmltopdf.app/Contents/MacOS/wkhtmltopdf", in, w)
	if err != nil {
		log.Println(err)
	}
}

type ServerConfig_Database struct {
	Driver         string `json:"driver"`
	DataSourceName string `json:"dsn"`
	PoolSize       int    `json:"poolSize"`
}
type ServerConfig struct {
	Database          ServerConfig_Database `json:"database"`
	ModelFile         string                `json:"modelFile"`
	TemplateRoot      string                `json:"templateRoot"`
	WebRoot           string                `json:"webRoot"`
	BindAddress       string                `json:"bindAddress"`
	PublicPatternsRaw []string              `json:"publicPatterns"`
}

func Serve(config *ServerConfig) {

	log.Println(config)
	parser := torch.Parser{
		Store:          torch.InMemorySessionStore(),
		Bath:           databath.RunABath(config.Database.Driver, config.Database.DataSourceName, config.Database.PoolSize),
		PublicPatterns: make([]*regexp.Regexp, len(config.PublicPatternsRaw), len(config.PublicPatternsRaw)),
	}

	for i, pattern := range config.PublicPatternsRaw {
		reg := regexp.MustCompile(pattern)
		parser.PublicPatterns[i] = reg
	}

	viewManager := GetViewManager(config.TemplateRoot)

	model, err := databath.ReadModelFromFile(config.ModelFile)
	if err != nil {
		panic("COULD NOT READ MODEL :" + err.Error())
	}

	socketManager := socket.GetManager(parser.Store)

	getHandler := SelectQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("get", &getHandler)

	setHandler := UpdateQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("set", &setHandler)

	createHandler := CreateQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("create", &createHandler)

	deleteHandler := DeleteQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("delete", &deleteHandler)

	choicesForHandler := ChoicesForQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("getChoicesFor", &choicesForHandler)

	customHandler := CustomQuery{
		Model: model,
		Bath:  parser.Bath,
	}
	socketManager.RegisterHandler("custom", &customHandler)

	loginViewHandler := ViewHandler{
		Manager:      viewManager,
		TemplateName: "login.html",
		Data:         nil,
	}

	signupViewHandler := ViewHandler{
		Manager:      viewManager,
		TemplateName: "signup.html",
		Data:         nil,
	}

	fallthroughHandler := GetFallthroughHandler(config)

	http.HandleFunc("/check", parser.Wrap(checkHandle))
	http.HandleFunc("/login", parser.WrapSplit(loginViewHandler.Handle, torch.HandleLogin))
	http.HandleFunc("/signup", parser.WrapSplit(signupViewHandler.Handle, signupHandle))
	http.HandleFunc("/pdf", parser.Wrap(pdfHandle))
	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/", parser.Wrap(fallthroughHandler.Handle))

	err = http.ListenAndServe(config.BindAddress, nil)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Server Stopped")
}
