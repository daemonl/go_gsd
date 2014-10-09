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
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/google_auth"
)

var re_unsafe *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9]`)

// Serve starts up a http server with all of the glorious configuration options
func Serve(config *ServerConfig) error {

	core, err := config.GetCore()
	if err != nil {
		return err
	}

	model := core.GetModel()

	lilo := torch.GetBasicLoginLogout(core.DB)

	sessionStore := torch.InMemorySessionStore(config.SessionDumpFile, lilo.LoadUserById, core.OpenDatabaseConnection)

	parser := torch.BasicParser(sessionStore, config.PublicPatternsRaw)

	// Register the AJAX and Socket methods (These methods work on both)
	handlerMap := map[string]actions.Handler{
		"get":     &actions.SelectQuery{Core: core},
		"set":     &actions.UpdateQuery{Core: core},
		"create":  &actions.CreateQuery{Core: core},
		"delete":  &actions.DeleteQuery{Core: core},
		"custom":  &actions.CustomQuery{Core: core},
		"dynamic": &actions.DynamicHandler{Core: core},
		"ping":    &actions.PingAction{Core: core},
		"msg":     &actions.MsgAction{Core: core},
	}

	socketManager := socket.GetManager(parser.Store)
	routes := router.GetRouter(parser, config.PublicPatternsRaw)

	for funcName, handler := range handlerMap {
		func(funcName string, handler actions.Handler) {
			socketManager.RegisterHandler(funcName, handler)
			routes.AddRoute("/ajax/"+funcName, actions.AsRouterHandler(handler), "POST")
		}(funcName, handler)
	}

	loginViewHandler := core.TemplateManager.GetSimpleTemplateHandler("login.html")
	setPasswordViewHandler := core.TemplateManager.GetSimpleTemplateHandler("set_password.html")

	fileHandler := file.GetFileHandler(config.UploadDirectory, model)

	if config.OAuthConfig != nil {
		oauthHandler := google_auth.OAuthHandler{
			Config:      config.OAuthConfig,
			LoginLogout: lilo,
		}
		http.HandleFunc("/oauth/return", parser.Wrap(oauthHandler.OauthResponse))
		http.HandleFunc("/oauth/request", parser.Wrap(oauthHandler.OauthRequest))
	}

	// SET UP URLS

	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/script/", core.runScript)

	if config.DevMode {
		log.Println("---DEV MODE---")
		routes.AddRoute("/app.html", &minihandlers.RawFileHandler{WebRoot: config.WebRoot, Filename: "app_dev.html"})
		http.Handle("/main.css", &minihandlers.LessHandler{WebRoot: config.WebRoot, Filename: "less/main.less"})
		http.Handle("/pdf.css", &minihandlers.LessHandler{WebRoot: config.WebRoot, Filename: "less/pdf.less"})
	}

	routes.AddRoute("/login", loginViewHandler, "GET")
	routes.AddRouteFunc("/login", lilo.HandleLogin, "POST")
	routes.AddRouteFunc("/logout", lilo.HandleLogout)
	routes.AddRoute("/set_password", setPasswordViewHandler, "GET")
	routes.AddRouteFunc("/set_password", lilo.HandleSetPassword, "POST")

	routes.AddRoutePathFunc("/report_html/%s/%d", core.Reporter.Handle, "GET")
	routes.AddRoutePathFunc("/report_pdf/%s/%d", core.PDFHandler.Handle, "GET")

	routes.AddRoutePathFunc("/emailpreview/%s/%d", core.Reporter.Handle, "GET")
	routes.AddRoutePathFunc("/sendmail/%s/%d/%s/%s", core.MailHandler.Handle)

	routes.AddRoutePathFunc("/csv/%s", core.CSVHandler.Handle)

	http.HandleFunc("/upload/", parser.Wrap(fileHandler.Upload))
	http.HandleFunc("/download/", parser.Wrap(fileHandler.Download))

	routes.Redirect("/", "app.html")
	routes.Fallthrough(config.WebRoot)

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
