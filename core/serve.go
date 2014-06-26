package core

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"

	"github.com/daemonl/databath/sync"
	"github.com/daemonl/go_gsd/actions"
	"github.com/daemonl/go_gsd/csv"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/file"
	"github.com/daemonl/go_gsd/pdf"
	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_lib/google_auth"
)

var re_unsafe *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9]`)

func signupHandle(request shared.IRequest) {
	//username := request.PostValueString("username")
	password := request.PostValueString("password")
	hashStore := torch.HashPassword(password)
	request.WriteString(hashStore)
}

func Sync(config *ServerConfig, force bool) error {
	core, err := config.GetCore()
	if err != nil {
		return err
	}
	db, err := core.OpenDatabaseConnection(nil)
	if err != nil {
		return err
	}
	defer db.Close()
	sync.SyncDb(db, core.GetModel(), force)
	return nil

}

// Serve starts up a http server with all of the glorious configuration options
func Serve(config *ServerConfig) error {

	core, err := config.GetCore()
	if err != nil {
		return err
	}

	model := core.GetModel()

	lilo := torch.GetBasicLoginLogout(core.DB)

	sessionStore := torch.InMemorySessionStore(config.SessionDumpFile, lilo.LoadUserById, core.OpenDatabaseConnection)

	parser := &torch.Parser{
		Store:          sessionStore,
		PublicPatterns: make([]*regexp.Regexp, len(config.PublicPatternsRaw), len(config.PublicPatternsRaw)),
	}

	// Insert the regexes for all 'public' urls
	for i, pattern := range config.PublicPatternsRaw {
		reg := regexp.MustCompile(pattern)
		parser.PublicPatterns[i] = reg
	}

	// Register the AJAX and Socket methods (These methods work on both)
	handlerMap := map[string]actions.Handler{
		"get":     &actions.SelectQuery{Core: core},
		"set":     &actions.UpdateQuery{Core: core},
		"create":  &actions.CreateQuery{Core: core},
		"delete":  &actions.DeleteQuery{Core: core},
		"custom":  &actions.CustomQuery{Core: core},
		"dynamic": &actions.DynamicHandler{Core: core},
		"ping":    &actions.PingAction{Core: core},
	}

	socketManager := socket.GetManager(parser.Store)
	routes := router.GetRouter(parser)

	for funcName, handler := range handlerMap {
		func(funcName string, handler actions.Handler) {
			socketManager.RegisterHandler(funcName, handler)
			routes.AddRoute("/ajax/"+funcName, actions.AsRouterHandler(handler), "POST")
		}(funcName, handler)
	}

	parser.Store.SetBroadcast(socketManager.Broadcast)

	config.ViewManager = view.GetViewManager(config.TemplateRoot, config.TemplateIncludeRoot)

	templateWriter := &view.TemplateWriter{
		Model:       model,
		ViewManager: config.ViewManager,
		Runner:      core.Runner,
		DB:          core.DB,
	}

	loginViewHandler := &view.ViewHandler{
		Manager:      config.ViewManager,
		TemplateName: "login.html",
		Data:         nil,
	}

	setPasswordViewHandler := &view.ViewHandler{
		Manager:      config.ViewManager,
		TemplateName: "set_password.html",
		Data:         nil,
	}

	emailHandler, err := email.GetEmailHandler(config.SmtpConfig, config.EmailConfig, templateWriter)
	if err != nil {
		log.Panic(err)
	}
	core.Email = emailHandler

	pdfHandler, err := pdf.GetPDFHandler(*config.PDFBinary, config.PDFConfig, templateWriter)
	if err != nil {
		log.Panic(err)
	}
	fileHandler := file.GetFileHandler(config.UploadDirectory, model)
	csvHandler := csv.GetCsvHandler(model)

	if config.OAuthConfig != nil {
		oauthHandler := google_auth.OAuthHandler{
			Config: config.OAuthConfig,
		}
		http.HandleFunc("/oauth/return", parser.Wrap(oauthHandler.OauthResponse))
		http.HandleFunc("/oauth/request", parser.Wrap(oauthHandler.OauthRequest))
	}

	fallthroughHandler := GetFallthroughHandler(config)

	// SET UP URLS

	http.Handle("/socket", socketManager.GetListener())

	routes.AddRoute("/login", loginViewHandler, "GET")
	routes.AddRouteFunc("/login", lilo.HandleLogin, "POST")
	routes.AddRouteFunc("/logout", lilo.HandleLogout)
	routes.AddRoute("/set_password", setPasswordViewHandler, "GET")
	routes.AddRouteFunc("/set_password", lilo.HandleSetPassword, "POST")

	routes.AddRoutePathFunc("/report_html/%s/%d", pdfHandler.Preview)
	routes.AddRoutePathFunc("/report_pdf/%s/%d", pdfHandler.GetPDF)

	routes.AddRoutePathFunc("/emailpreview/", emailHandler.Preview)
	routes.AddRoutePathFunc("/sendmail/", emailHandler.Send)

	//http.HandleFunc("/report_html/", parser.Wrap(pdfHandler.Preview))
	//http.HandleFunc("/report_pdf/", parser.Wrap(pdfHandler.GetPdf))
	http.HandleFunc("/upload/", parser.Wrap(fileHandler.Upload))
	http.HandleFunc("/download/", parser.Wrap(fileHandler.Download))
	http.HandleFunc("/csv/", parser.Wrap(csvHandler.Handle))
	http.HandleFunc("/reload", parser.Wrap(config.ReloadHandle))

	http.HandleFunc("/script/", core.runScript)

	if config.DevMode {
		log.Println("---DEV MODE---")
		http.Handle("/app.html", &rawFileHandler{config: config, filename: "app_dev.html"})
		http.Handle("/main.css", &lessHandler{config: config, filename: "less/main.less"})
		http.Handle("/pdf.css", &lessHandler{config: config, filename: "less/pdf.less"})
	}

	routes.Fallthrough(parser.Wrap(fallthroughHandler.Handle))
	http.Handle("/", routes)

	// SERVE!

	go func() {
		err = http.ListenAndServe(config.BindAddress, nil)
		if err != nil {
			log.Println(err)
		}
	}()

	sigChan := make(chan os.Signal)

	signal.Notify(sigChan, os.Interrupt)

	for {
		sigVal := <-sigChan
		log.Printf("SIG %s\n", sigVal.String())
		if sigVal == os.Interrupt {
			break
		}
	}

	log.Printf("=== SHUTDOWN INITIATED ===")

	sessionStore.DumpSessions()

	log.Println("=== SHUTDOWN COMPLETE ===")

	return nil
}

type rawFileHandler struct {
	filename string
	config   *ServerConfig
}

func (h *rawFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open(h.config.WebRoot + "/" + h.filename)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}

type lessHandler struct {
	filename string
	config   *ServerConfig
}

func (h *lessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("LESS HANDLER")
	w.Header().Add("Content-Type", "text/css")
	c := exec.Command("lessc", h.config.WebRoot+"/"+h.filename)
	c.Stdout = w
	c.Stderr = os.Stderr
	c.Run()
}
