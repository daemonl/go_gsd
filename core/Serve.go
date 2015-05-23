package core

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	_ "github.com/go-sql-driver/mysql"

	"github.com/daemonl/go_gsd/actions"
	"github.com/daemonl/go_gsd/file"
	"github.com/daemonl/go_gsd/minihandlers"
	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
)

var re_unsafe *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9]`)

type Server struct {
	router       router.Router
	sessionStore shared.ISessionStore
	core         *GSDCore
	config       *ServerConfig
}

type IServer interface {
	Serve() error
	AddRoute(format string, handler shared.IHandler, methods ...string) error
	AddRouteFunc(format string, handlerFunc func(shared.IRequest) (shared.IResponse, error), methods ...string) error
	AddRoutePathFunc(format string, handlerPathFunc func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error
}

// Serve starts up a http server with all of the glorious configuration options
func Serve(config *ServerConfig) error {
	server, err := BuildServer(config)
	if err != nil {
		return err
	}
	return server.Serve()
}

func BuildServer(config *ServerConfig) (IServer, error) {

	s := &Server{
		config: config,
	}
	var err error

	s.core, err = config.GetCore()
	if err != nil {
		return nil, err
	}

	lilo := torch.GetBasicLoginLogout(s.core.DB, "staff")

	authData := map[string]interface{}{}
	if config.GoogleAuthConfig != nil {
		lilo.AddAuthenticator(config.GoogleAuthConfig)
		authData["AuthCallback"] = config.GoogleAuthConfig.RedirectURI
		authData["GoogleClientID"] = config.GoogleAuthConfig.ClientID
	}

	s.sessionStore = torch.InMemorySessionStore(config.SessionDumpFile, lilo.LoadUserById, s.core.OpenDatabaseConnection, s.core.CloseDatabaseConnection)

	parser := torch.BasicParser(s.sessionStore, config.PublicPatternsRaw)

	// Register the AJAX and Socket methods (These methods work on both)
	handlerMap := map[string]actions.Handler{
		"get":     &actions.SelectQuery{Core: s.core},
		"set":     &actions.UpdateQuery{Core: s.core},
		"create":  &actions.CreateQuery{Core: s.core},
		"delete":  &actions.DeleteQuery{Core: s.core},
		"custom":  &actions.CustomQuery{Core: s.core},
		"dynamic": &actions.DynamicHandler{Core: s.core},
		"ping":    &actions.PingAction{Core: s.core},
		"msg":     &actions.MsgAction{Core: s.core},
	}

	socketManager := socket.GetManager(parser.Store)
	s.router = router.GetRouter(parser, config.PublicPatternsRaw)

	for funcName, handler := range handlerMap {
		func(funcName string, handler actions.Handler) {
			socketManager.RegisterHandler(funcName, handler)
			s.router.AddRoute("/ajax/"+funcName, actions.AsRouterHandler(handler), "POST")
		}(funcName, handler)
	}

	loginViewHandler := s.core.TemplateManager.GetSimpleTemplateHandler("login.html", authData)

	setPasswordViewHandler := s.core.TemplateManager.GetSimpleTemplateHandler("set_password.html", nil)

	model := s.core.GetModel()
	fileHandler := file.GetFileHandler(config.UploadDirectory, model)

	/*
		oauthHandler := google_auth.OAuthHandler{
			Config:      config.OAuthConfig,
			LoginLogout: lilo,
		}
		http.HandleFunc("/oauth/return", parser.Wrap(oauthHandler.OauthResponse))
		http.HandleFunc("/oauth/request", parser.Wrap(oauthHandler.OauthRequest))
	}*/

	// SET UP URLS

	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/script/", s.core.runScript)

	if config.DevMode {
		log.Println("---DEV MODE---")
		s.router.AddRoute("/app.html", &minihandlers.RawFileHandler{WebRoot: config.WebRoot, Filename: "app_dev.html"})
		http.Handle("/main.css", &minihandlers.LessHandler{WebRoot: config.WebRoot, Filename: "less/main.less"})
		http.Handle("/pdf.css", &minihandlers.LessHandler{WebRoot: config.WebRoot, Filename: "less/pdf.less"})
	}

	s.router.AddRoute("/login", loginViewHandler, "GET")
	s.router.AddRouteFunc("/oauth_callback", lilo.HandleOauthCallback, "GET")
	s.router.AddRouteFunc("/login", lilo.HandleLogin, "POST")
	s.router.AddRouteFunc("/logout", lilo.HandleLogout)
	s.router.AddRoute("/set_password", setPasswordViewHandler, "GET")
	s.router.AddRouteFunc("/set_password", lilo.HandleSetPassword, "POST")

	s.router.AddRoutePathFunc("/report/%s/%d/*.html", s.core.Reporter.Handle, "GET")
	s.router.AddRoutePathFunc("/report/%s/%d/*.pdf", s.core.PDFHandler.Handle, "GET")

	s.router.AddRoutePathFunc("/emailpreview/%s/%d", s.core.Reporter.Handle, "GET")
	s.router.AddRoutePathFunc("/sendmail/%s/%d/%s/%s", s.core.MailHandler.Handle)

	s.router.AddRoutePathFunc("/csv/%s", s.core.CSVHandler.Handle)

	http.HandleFunc("/upload/", parser.Wrap(fileHandler.Upload))
	http.HandleFunc("/download/", parser.Wrap(fileHandler.Download))

	s.router.Redirect("/", "app.html")
	s.router.Fallthrough(config.WebRoot)

	return s, nil
}

func (s *Server) AddRoute(format string, handler shared.IHandler, methods ...string) error {
	return s.router.AddRoute(format, handler, methods...)
}

func (s *Server) AddRouteFunc(format string, hf func(shared.IRequest) (shared.IResponse, error), methods ...string) error {
	return s.router.AddRouteFunc(format, hf, methods...)
}

func (s *Server) AddRoutePathFunc(format string, pathRequestFunc func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error {
	return s.router.AddRoutePathFunc(format, pathRequestFunc, methods...)
}

func (s *Server) Serve() error {
	var err error

	http.Handle("/", s.router)

	// SERVE!

	go func() {
		err = http.ListenAndServe(s.config.BindAddress, nil)
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

	s.sessionStore.DumpSessions()

	log.Println("=== SHUTDOWN COMPLETE ===")

	return nil
}
