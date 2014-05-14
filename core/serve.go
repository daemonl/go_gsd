package core

import (
	"encoding/json"
	"github.com/daemonl/go_gsd/actions"
	"github.com/daemonl/go_gsd/csv"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/file"
	"github.com/daemonl/go_gsd/pdf"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_lib/databath"
	"github.com/daemonl/go_lib/databath/sync"
	"github.com/daemonl/go_lib/google_auth"

	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
)

var re_unsafe *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9]`)

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

type ServerConfig_Database struct {
	Driver         string `json:"driver"`
	DataSourceName string `json:"dsn"`
	PoolSize       int    `json:"poolSize"`
}
type ServerConfig struct {
	Database            ServerConfig_Database `json:"database"`
	ModelFile           string                `json:"modelFile"`
	TemplateRoot        string                `json:"templateRoot"`
	WebRoot             string                `json:"webRoot"`
	BindAddress         string                `json:"bindAddress"`
	PublicPatternsRaw   []string              `json:"publicPatterns"`
	UploadDirectory     string                `json:"uploadDirectory"`
	TemplateIncludeRoot string                `json:"templateIncludeRoot"`
	ScriptDirectory     string                `json:"scriptDirectory"`

	EmailConfig     *email.EmailHandlerConfig
	EmailFile       *string           `json:"emailFile"`
	SmtpConfig      *email.SmtpConfig `json:"smtpConfig"`
	SessionDumpFile *string           `json:"sessionDumpFile"`

	PdfConfig *pdf.PdfHandlerConfig
	PdfFile   *string `json:"pdfFile"`
	PdfBinary *string `json:"pdfBinary"`

	OAuthConfig *google_auth.OAuthConfig `json:"oauth"`

	ViewManager *view.ViewManager
}

func (config *ServerConfig) ReloadHandle(requestTorch *torch.Request) {
	err := config.ViewManager.Reload()
	if err != nil {
		requestTorch.Writef("Error Loading Views: %s", err.Error())
		return
	}
	requestTorch.Writef("Loaded Views")
}

func GetCore(config *ServerConfig) (core *GSDCore, err error) {

	bath := databath.RunABath(config.Database.Driver, config.Database.DataSourceName, config.Database.PoolSize)

	model, err := databath.ReadModelFromFile(config.ModelFile)
	if err != nil {
		log.Println("COULD NOT READ MODEL FILE")
		log.Println(err.Error())
		return nil, err
	}

	core = &GSDCore{
		Bath:   bath,
		Model:  model,
		Config: config,
	}

	core.Runner = &dynamic.DynamicRunner{
		DataBath:      core.Bath,
		BaseDirectory: core.Config.ScriptDirectory, // "/home/daemonl/schkit/impl/pov/script/",
		SendMail:      core.SendMail,
	}

	core.Hooker = &Hooker{
		Core: core,
	}

	return

}

func Sync(config *ServerConfig, force bool) error {
	core, err := GetCore(config)
	if err != nil {
		return err
	}

	conn := core.Bath.GetConnection()
	db := conn.GetDB()
	sync.SyncDb(db, core.GetModel(), force)
	return nil

}
func Serve(config *ServerConfig) error {

	core, err := GetCore(config)
	if err != nil {
		return err
	}

	model := core.GetModel()

	parser := torch.Parser{
		Store:          torch.InMemorySessionStore(),
		Bath:           core.Bath,
		PublicPatterns: make([]*regexp.Regexp, len(config.PublicPatternsRaw), len(config.PublicPatternsRaw)),
	}

	if config.SessionDumpFile != nil {
		log.Println("-\n========\nHIDRATE SESSIONS\n========")

		sessFile, err := os.Open(*config.SessionDumpFile)
		if err != nil {
			log.Printf("Could not load sessions: %s\n", err.Error())
		} else {
			conn := core.Bath.GetConnection()
			db := conn.GetDB()
			parser.Store.LoadSessions(sessFile, func(id uint64) (*torch.User, error) {
				return torch.LoadUserById(db, id)
			})
			conn.Release()
			sessFile.Close()
		}
		log.Println("-\n========\nEND SESSIONS\n========")
	}

	for i, pattern := range config.PublicPatternsRaw {
		reg := regexp.MustCompile(pattern)
		parser.PublicPatterns[i] = reg
	}

	actionMap := map[string]socket.Handler{
		"get":     &actions.SelectQuery{Core: core},
		"set":     &actions.UpdateQuery{Core: core},
		"create":  &actions.CreateQuery{Core: core},
		"delete":  &actions.DeleteQuery{Core: core},
		"custom":  &actions.CustomQuery{Core: core},
		"dynamic": &actions.DynamicHandler{Core: core},
		"ping":    &actions.PingAction{Core: core},
	}

	socketManager := socket.GetManager(parser.Store)

	for funcName, handler := range actionMap {

		func(funcName string, handler socket.Handler) {
			socketManager.RegisterHandler(funcName, handler)

			rFunc := func(request *torch.Request) {

				requestObject := handler.GetRequestObject()
				writer, r := request.GetRaw()
				dec := json.NewDecoder(r.Body)
				err := dec.Decode(requestObject)
				if err != nil {
					request.DoError(err)
					return
				}
				responseObject, err := handler.HandleRequest(request, requestObject)
				if err != nil {
					request.DoError(err)
					return
				}
				enc := json.NewEncoder(writer)
				enc.Encode(responseObject)
			}

			http.HandleFunc("/ajax/"+funcName, parser.Wrap(rFunc))
		}(funcName, handler)
	}

	parser.Store.Broadcast = socketManager.Broadcast

	config.ViewManager = view.GetViewManager(config.TemplateRoot, config.TemplateIncludeRoot)

	templateWriter := view.TemplateWriter{
		Bath:        core.Bath,
		Model:       model,
		ViewManager: config.ViewManager,
		Runner:      core.Runner,
	}

	loginViewHandler := view.ViewHandler{
		Manager:      config.ViewManager,
		TemplateName: "login.html",
		Data:         nil,
	}

	signupViewHandler := view.ViewHandler{
		Manager:      config.ViewManager,
		TemplateName: "signup.html",
		Data:         nil,
	}

	setPasswordViewHandler := view.ViewHandler{
		Manager:      config.ViewManager,
		TemplateName: "set_password.html",
		Data:         nil,
	}

	emailHandler, err := email.GetEmailHandler(config.SmtpConfig, config.EmailConfig, &templateWriter)
	if err != nil {
		log.Panic(err)
	}
	core.Email = emailHandler

	pdfHandler, err := pdf.GetPdfHandler(*config.PdfBinary, config.PdfConfig, &templateWriter)
	if err != nil {
		log.Panic(err)
	}
	fileHandler := file.GetFileHandler(config.UploadDirectory, core.Bath, model)
	csvHandler := csv.GetCsvHandler(core.Bath, model)

	if config.OAuthConfig != nil {
		oauthHandler := google_auth.OAuthHandler{
			Config: config.OAuthConfig,
		}
		http.HandleFunc("/oauth/return", parser.Wrap(oauthHandler.OauthResponse))
		http.HandleFunc("/oauth/request", parser.Wrap(oauthHandler.OauthRequest))
	}

	fallthroughHandler := GetFallthroughHandler(config)

	runScript := func(w http.ResponseWriter, r *http.Request) {
		// This should only run for requests internal to the system, for cronjobs and scripts.
		// Nginx sets x-real-ip on all forwarded requests.
		// This needs a security review.

		real_ip := r.Header.Get("X-Real-IP")
		if len(real_ip) > 0 {
			w.Write([]byte("NOT ALLOWED"))
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			log.Printf("/script/ called with %s\n", parts)
			return
		}
		script := re_unsafe.ReplaceAllString(parts[2], "_")

		log.Printf("RUN SCRIPT %s\n", script)

		_, err := core.Runner.Run(script+".js", map[string]interface{}{"path": parts[2:]})
		if err != nil {
			log.Println(err)
			w.Write([]byte("ERROR"))
			return
		}
		w.Write([]byte("OK"))
	}

	// SET UP URLS

	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/check", parser.Wrap(checkHandle))
	http.HandleFunc("/login", parser.WrapSplit(loginViewHandler.Handle, torch.HandleLogin))
	http.HandleFunc("/logout", parser.Wrap(torch.HandleLogout))
	http.HandleFunc("/set_password", parser.WrapSplit(setPasswordViewHandler.Handle, torch.HandleSetPassword))
	http.HandleFunc("/signup", parser.WrapSplit(signupViewHandler.Handle, signupHandle))
	http.HandleFunc("/report_html/", parser.Wrap(pdfHandler.Preview))
	http.HandleFunc("/report_pdf/", parser.Wrap(pdfHandler.GetPdf))
	http.HandleFunc("/upload/", parser.Wrap(fileHandler.Upload))
	http.HandleFunc("/download/", parser.Wrap(fileHandler.Download))
	http.HandleFunc("/csv/", parser.Wrap(csvHandler.Handle))
	http.HandleFunc("/reload", parser.Wrap(config.ReloadHandle))
	http.HandleFunc("/emailpreview/", parser.Wrap(emailHandler.Preview))
	http.HandleFunc("/sendmail/", parser.Wrap(emailHandler.Send))
	http.HandleFunc("/script/", runScript)
	http.HandleFunc("/", parser.Wrap(fallthroughHandler.Handle))

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

	log.Printf("= SHUTDOWN INITIATED")

	if config.SessionDumpFile != nil {
		log.Println("-\n========\nDUMP SESSIONS\n========")

		sessFile, err := os.Create(*config.SessionDumpFile)
		if err != nil {
			log.Printf("Could not save sessions: %s\n", err.Error())
		} else {
			parser.Store.DumpSessions(sessFile)
			sessFile.Close()
		}
	}

	log.Println("= SHUTDOWN COMPLETE ")

	return nil
}
