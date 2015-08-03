package hooker

import (
	"log"
	"time"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/components"
)

type Hooker struct {
	Model    *databath.Model
	Runner   components.Runner
	Reporter components.Reporter
	Mailer   components.Mailer
}

func (h *Hooker) DoPreHooks(hc *components.HookContext) {
	log.Println("PROCESS PRE HOOKS")
	collection := h.Model.Collections[hc.ActionSummary.Collection]
	for _, hook := range collection.Hooks {
		if hook.Applies(hc) {
			hook.RunPreHook(hc)
		}
	}
}

func (h *Hooker) DoPostHooks(hc *components.HookContext) {
	log.Println("PROCESS POST HOOKS")
	go h.WriteHistory(hc)
	collection := h.Model.Collections[hc.ActionSummary.Collection]
	for _, hook := range collection.Hooks {
		if hook.Applies(hc) {
			hook.RunPostHook(hc)
		}
	}
}

func (h *Hooker) WriteHistory(hc *components.HookContext) {

	identity, _ := h.Model.GetIdentityString(hc.DB, hc.ActionSummary.Collection, hc.ActionSummary.Pk)
	timestamp := time.Now().Unix()

	//log.Println("WRITE HISTORY", as.UserId, identity, timestamp, as.Action, as.Collection, as.Pk)

	_, err := hc.DB.Exec(
		`INSERT INTO history 
		(user, identity, timestamp, action, entity, entity_id) VALUES 
		(?, ?, ?, ?, ?, ?)
		`, hc.ActionSummary.UserId, identity, timestamp, hc.ActionSummary.Action, hc.ActionSummary.Collection, hc.ActionSummary.Pk)

	if err != nil {
		log.Println(err)
		return
	}
}
