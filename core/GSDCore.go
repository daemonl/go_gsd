package core

import (
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/shared_structs"
	"github.com/daemonl/go_lib/databath"
	"log"
)

type GSDCore struct {
	Bath   *databath.Bath
	Model  *databath.Model
	Email  *email.EmailHandler
	Hooker *Hooker
	Config *ServerConfig
	Runner *dynamic.DynamicRunner
}

func (core *GSDCore) GetConnection() *databath.Connection {
	return core.Bath.GetConnection()
}

func (core *GSDCore) DoHooksPreAction(as *shared_structs.ActionSummary) {
	core.Hooker.DoPreHooks(as)
}

func (core *GSDCore) DoHooksPostAction(as *shared_structs.ActionSummary) {
	core.Hooker.DoPostHooks(as)
}

func (core *GSDCore) GetModel() *databath.Model {
	return core.Model
}

func (core *GSDCore) RunDynamic(filename string, parameters map[string]interface{}) (map[string]interface{}, error) {
	return core.Runner.Run(filename, parameters)
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
