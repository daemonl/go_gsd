package quickserve

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/go-sql-driver/mysql"

	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
)

type Server struct {
	Config          *ServerConfig
	DB              *sql.DB
	router          router.Router
	sessionStore    shared.ISessionStore
	parser          *torch.Parser
	TemplateManager *view.TemplateManager
}

var configFilename string
var db *sql.DB

func (s *Server) getDb(shared.ISession) (*sql.DB, error) {
	return s.DB, nil
}

func GetServer(config *ServerConfig) (*Server, error) {

	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		return nil, err
	}

	config.PublicPatterns = append(config.PublicPatterns, "/login")

	s := &Server{
		Config: config,
		DB:     db,
	}

	templateManager := view.GetTemplateManager(config.TemplateRoot, config.TemplateIncludeRoot)

	s.TemplateManager = templateManager

	lilo := torch.GetBasicLoginLogout(db, "user")
	s.sessionStore = torch.InMemorySessionStore(config.SessionDumpFile, lilo.LoadUserById, s.getDb)
	s.parser = torch.BasicParser(s.sessionStore, config.PublicPatterns)
	s.router = router.GetRouter(s.parser, config.PublicPatterns)

	s.router.Fallthrough(config.PublicRoot)

	s.router.AddRouteFunc("/login", lilo.HandleLogin, "POST")
	s.router.AddRouteFunc("/logout", lilo.HandleLogout)
	s.router.AddRouteFunc("/set_password", lilo.HandleSetPassword, "POST")
	return s, nil

}

func (s *Server) AddRoute(path string, handler func(shared.IPathRequest) (shared.IResponse, error), methods ...string) error {
	return s.router.AddRoutePathFunc(path, handler, methods...)
}

func (s *Server) Serve() error {
	http.Handle("/", s.router)

	go func() {
		err := http.ListenAndServe(s.Config.BindAddress, nil)
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

	s.sessionStore.DumpSessions()

	return nil
}
