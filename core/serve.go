package core

import (
	"github.com/daemonl/go_gsd/csv"
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
)

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

func Serve(config *ServerConfig) {

	bath := databath.RunABath(config.Database.Driver, config.Database.DataSourceName, config.Database.PoolSize)

	model, err := databath.ReadModelFromFile(config.ModelFile)
	if err != nil {
		panic("COULD NOT READ MODEL :" + err.Error())
	}

	// IF IT IS TO SYNC, STOP HERE (--sync)
	if doSync {
		conn := bath.GetConnection()
		db := conn.GetDB()
		sync.SyncDb(db, model, forceSync)
		return
	}

	core := GSDCore{
		Bath:   bath,
		Model:  model,
		Config: config,
	}
	h := Hooker{
		Core: &core,
	}

	core.Hooker = &h

	parser := torch.Parser{
		Store:          torch.InMemorySessionStore(),
		Bath:           bath,
		PublicPatterns: make([]*regexp.Regexp, len(config.PublicPatternsRaw), len(config.PublicPatternsRaw)),
	}

	if config.SessionDumpFile != nil {
		log.Println("-\n========\nHIDRATE SESSIONS\n========")

		sessFile, err := os.Open(*config.SessionDumpFile)
		if err != nil {
			log.Printf("Could not load sessions: %s\n", err.Error())
		} else {
			conn := bath.GetConnection()
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

	config.ViewManager = view.GetViewManager(config.TemplateRoot, config.TemplateIncludeRoot)

	templateWriter := view.TemplateWriter{
		Bath:        bath,
		Model:       model,
		ViewManager: config.ViewManager,
	}

	socketManager := socket.GetManager(parser.Store)

	getHandler := SelectQuery{Core: &core}
	socketManager.RegisterHandler("get", &getHandler)

	setHandler := UpdateQuery{Core: &core}
	socketManager.RegisterHandler("set", &setHandler)

	createHandler := CreateQuery{Core: &core}
	socketManager.RegisterHandler("create", &createHandler)

	deleteHandler := DeleteQuery{Core: &core}
	socketManager.RegisterHandler("delete", &deleteHandler)

	choicesForHandler := ChoicesForQuery{Core: &core}
	socketManager.RegisterHandler("getChoicesFor", &choicesForHandler)

	customHandler := CustomQuery{Core: &core}
	socketManager.RegisterHandler("custom", &customHandler)

	pingHandler := PingHandler{}
	socketManager.RegisterHandler("ping", &pingHandler)

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
	fileHandler := file.GetFileHandler(config.UploadDirectory, parser.Bath, model)
	csvHandler := csv.GetCsvHandler(parser.Bath, model)

	if config.OAuthConfig != nil {
		oauthHandler := google_auth.OAuthHandler{
			Config: config.OAuthConfig,
		}
		http.HandleFunc("/oauth/return", parser.Wrap(oauthHandler.OauthResponse))
		http.HandleFunc("/oauth/request", parser.Wrap(oauthHandler.OauthRequest))
	}

	fallthroughHandler := GetFallthroughHandler(config)

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
	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/", parser.Wrap(fallthroughHandler.Handle))

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
}
