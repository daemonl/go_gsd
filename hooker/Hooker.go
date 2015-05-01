package hooker

import (
	"database/sql"
	"log"
	"time"

	"github.com/daemonl/go_gsd/shared"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/components"
)

type Hooker struct {
	Model    *databath.Model
	Runner   components.Runner
	Reporter components.Reporter
	Mailer   components.Mailer
}

func (h *Hooker) DoPreHooks(db *sql.DB, as *shared.ActionSummary, session shared.ISession) {

	model := h.Model
	collection := model.Collections[as.Collection]

	for _, hook := range collection.Hooks {
		if hook.When.What != as.Action {
			continue
		}
		log.Println("HOOK: " + hook.Collection)

		_, ok := as.Fields[hook.When.Field]
		if !ok {
			continue
		}

		// WOOT, Hook Matches. Let's Do this shit.

		// Add all fields to the update.

		for k, v := range hook.Set {
			_, exists := as.Fields[k]
			if exists {
				continue
			}
			vString, ok := v.(string)
			if ok {
				if vString == "#me" {
					v = as.UserId
				}
			}
			as.Fields[k] = v
		}

	}
}
func (h *Hooker) DoPostHooks(db *sql.DB, as *shared.ActionSummary, session shared.ISession) {
	go h.WriteHistory(db, as, session)

	log.Println("PROCESS POST HOOKS")

	model := h.Model
	collection := model.Collections[as.Collection]

	for _, hook := range collection.Hooks {
		if hook.CustomAction != nil {
			log.Println("HOOK CUSTOM ACTION: " + hook.Collection)
			p := make([]interface{}, len(hook.Raw.InFields), len(hook.Raw.InFields))
			for i, rawField := range hook.Raw.InFields {
				val, ok := rawField["val"]
				if !ok {
					log.Println("No 'val' in custom query hook value")
					return
				}
				str, ok := val.(string)
				if ok {
					if str == "#id" {
						val = as.Pk
					}
				}
				p[i] = val
			}
			results, err := hook.CustomAction.Run(db, p)
			if err != nil {
				log.Println(err.Error())
				return
			}
			log.Println(results)
		}
		for _, scriptName := range hook.Scripts {

			log.Printf("Hook Script %s\n", scriptName)

			scriptMap := map[string]interface{}{
				"userId":     as.UserId,
				"action":     as.Action,
				"collection": as.Collection,
				"id":         as.Pk,
				"fields":     as.Fields,
			}

			dr := h.Runner

			fnConfig, ok := h.Model.DynamicFunctions[scriptName]
			if !ok {
				log.Printf("No registered dynamic function named '%s'", scriptName)
				return
			}

			_, err := dr.Run(fnConfig.Filename, scriptMap, db)
			if err != nil {
				log.Println(err.Error())
				return
			}
			log.Println("Hook script complete")

		}
		if hook.Email != nil {

			if hook.When.What != as.Action {
				continue
			}
			log.Println("HOOK: " + hook.Collection)

			_, ok := as.Fields[hook.When.Field]
			if !ok {
				continue
			}

			// WOOT, Hook Matches. Let's Do this shit.
			log.Println("Send Email " + hook.Email.Template + " TO " + hook.Email.Recipient)

			//hook.Email.Template, as.Pk,
			//viewData :=
			log.Printf("TPL: %s\n", hook.Email.Template)
			report, err := h.Reporter.GetReportHTMLWriter(hook.Email.Template, as.Pk, session)
			if err != nil {
				log.Println(err.Error())
				return
			}

			go h.Mailer.SendResponse(report, hook.Email.Recipient, "")

		}
	}
}

func (h *Hooker) WriteHistory(db *sql.DB, as *shared.ActionSummary, session shared.ISession) {
	//, userId uint64, action string, collectionName string, entityId uint64) {

	identity, _ := h.Model.GetIdentityString(db, as.Collection, as.Pk)
	timestamp := time.Now().Unix()

	log.Println("WRITE HISTORY", as.UserId, identity, timestamp, as.Action, as.Collection, as.Pk)

	_, err := db.Exec(
		`INSERT INTO history 
		(user, identity, timestamp, action, entity, entity_id) VALUES 
		(?, ?, ?, ?, ?, ?)
		`, as.UserId, identity, timestamp, as.Action, as.Collection, as.Pk)

	if err != nil {
		log.Println(err)
		return
	}
}
