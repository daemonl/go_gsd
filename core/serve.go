package core

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/daemonl/databath/sync"
	"github.com/daemonl/go_gsd/actions"
	"github.com/daemonl/go_gsd/csv"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/file"
	"github.com/daemonl/go_gsd/pdf"
	"github.com/daemonl/go_gsd/socket"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_lib/google_auth"
)

var re_unsafe *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9]`)

func signupHandle(request torch.Request) {
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

	parser := torch.Parser{
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

	for funcName, handler := range handlerMap {

		func(funcName string, handler actions.Handler) {

			// Register handler with the Socket Manager
			socketManager.RegisterHandler(funcName, handler)

			// Wrap the action in a torch request
			rFunc := func(request torch.Request) {

				requestData := handler.RequestDataPlaceholder()
				writer, r := request.GetRaw()

				// Decode request body JSON to object
				dec := json.NewDecoder(r.Body)
				err := dec.Decode(requestData)
				if err != nil {
					request.DoError(err)
					return
				}

				// Do the request
				responseObject, err := handler.HandleRequest(request, requestData)

				// Encode response object as JSON
				if err != nil {
					request.DoError(err)
					return
				}
				enc := json.NewEncoder(writer)
				enc.Encode(responseObject)
			}

			// Add the wrapped torch request to the standard handler.
			http.HandleFunc("/ajax/"+funcName, parser.Wrap(rFunc))
		}(funcName, handler)
	}

	parser.Store.SetBroadcast(socketManager.Broadcast)

	config.ViewManager = view.GetViewManager(config.TemplateRoot, config.TemplateIncludeRoot)

	templateWriter := view.TemplateWriter{
		Model:       model,
		ViewManager: config.ViewManager,
		Runner:      core.Runner,
		DB:          core.DB,
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

		db, err := core.UsersDatabase()
		if err != nil {
			log.Println(err)
			w.Write([]byte("ERROR"))
			return
		}
		_, err = core.Runner.Run(script+".js", map[string]interface{}{"path": parts[2:]}, db)
		if err != nil {
			log.Println(err)
			w.Write([]byte("ERROR"))
			return
		}
		w.Write([]byte("OK"))
	}

	// SET UP URLS

	http.Handle("/socket", socketManager.GetListener())
	http.HandleFunc("/login", parser.WrapSplit(loginViewHandler.Handle, lilo.HandleLogin))
	http.HandleFunc("/logout", parser.Wrap(lilo.HandleLogout))
	http.HandleFunc("/set_password", parser.WrapSplit(setPasswordViewHandler.Handle, lilo.HandleSetPassword))
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

	if config.DevMode {
		log.Println("---DEV MODE---")
		http.Handle("/app.html", &rawFileHandler{config: config, filename: "app_dev.html"})
		http.Handle("/main.css", &lessHandler{config: config, filename: "less/main.less"})
		http.Handle("/pdf.css", &lessHandler{config: config, filename: "less/pdf.less"})
	}

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
