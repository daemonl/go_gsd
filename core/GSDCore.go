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

	components.Hooker
	components.Mailer
	components.Reporter
	components.PDFer
	components.Runner
	components.Xero

	CSVHandler  shared.IPathHandler
	PDFHandler  shared.IPathHandler
	MailHandler shared.IPathHandler

	TemplateManager *view.TemplateManager
}

func (core *GSDCore) OpenDatabaseConnection(session shared.ISession) (*sql.DB, error) {
	return core.DB, nil
	//return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) CloseDatabaseConnection(db *sql.DB) {
	// leave it. open.
}

func (core *GSDCore) UsersDatabase() (*sql.DB, error) {
	return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) GetModel() *databath.Model {
	return core.Model
}
