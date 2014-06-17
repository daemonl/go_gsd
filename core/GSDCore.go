package core

import (
	"database/sql"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	"log"
)

type GSDCore struct {
	Model  *databath.Model
	Email  *email.EmailHandler
	Hooker *Hooker
	Config *ServerConfig
	Runner *dynamic.DynamicRunner
}

func (core *GSDCore) DB(session *torch.Session) (*sql.DB, error) {
	return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) UsersDatabase() (*sql.DB, error) {
	return sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
}

func (core *GSDCore) DoHooksPreAction(db *sql.DB, as *shared_structs.ActionSummary) {
	core.Hooker.DoPreHooks(db, as)
}

func (core *GSDCore) DoHooksPostAction(db *sql.DB, as *shared_structs.ActionSummary) {
	core.Hooker.DoPostHooks(db, as)
}

func (core *GSDCore) GetModel() *databath.Model {
	return core.Model
}

func (core *GSDCore) RunDynamic(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error) {
	return core.Runner.Run(filename, parameters, db)
}

func (core *GSDCore) SendMail(to string, subject string, body string) {
	log.Printf("SEND MAIL TO %s: %s\n", to, subject)
	e := &email.Email{
		Recipient: to,
		Sender:    core.Config.EmailConfig.From,
		Subject:   subject,
		Html:      body,
	}
	go core.Email.Sender.Send(e)
}
