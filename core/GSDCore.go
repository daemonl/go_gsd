package core

import (
	"database/sql"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/components"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/view"
)

type GSDCore struct {
	Config *ServerConfig
	Model  *databath.Model

	DB *sql.DB

	Hooker   components.Hooker
	Mailer   components.Mailer
	Reporter components.Reporter
	PDFer    components.PDFer
	Runner   components.Runner

	CSVHandler  shared.IPathHandler
	PDFHandler  shared.IPathHandler
	MailHandler shared.IPathHandler

	TemplateManager *view.TemplateManager
}

func (core *GSDCore) OpenDatabaseConnection(session shared.ISession) (*sql.DB, error) {
	return core.DB, nil
	//return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) UsersDatabase() (*sql.DB, error) {
	return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) DoHooksPreAction(db *sql.DB, as *shared.ActionSummary, session shared.ISession) {
	core.Hooker.DoPreHooks(db, as, session)
}

func (core *GSDCore) DoHooksPostAction(db *sql.DB, as *shared.ActionSummary, session shared.ISession) {
	core.Hooker.DoPostHooks(db, as, session)
}

func (core *GSDCore) GetModel() *databath.Model {
	return core.Model
}

func (core *GSDCore) RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error) {
	return core.Runner.Run(filename, parameters, db)
}

func (core *GSDCore) SendMail(to string, subject string, body string) {
	core.Mailer.SendSimple(to, subject, body)
}
