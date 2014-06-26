package core

import (
	"database/sql"
	"fmt"
	"github.com/daemonl/go_gsd/shared"
	"log"
	"time"
)

type Hooker struct {
	Core *GSDCore
}

func (h *Hooker) DoPreHooks(db *sql.DB, as *shared.ActionSummary) {

	model := h.Core.Model
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
func (h *Hooker) DoPostHooks(db *sql.DB, as *shared.ActionSummary) {
	go h.WriteHistory(db, as)

	log.Println("PROCESS POST HOOKS")

	model := h.Core.Model
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

			dr := h.Core.Runner

			fnConfig, ok := h.Core.Model.DynamicFunctions[scriptName]
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
			report, err := h.Core.Email.GetReport(hook.Email.Template, as.Pk, nil)
			if err != nil {
				log.Println(err.Error())
				return
			}

			viewData, err := report.PrepareData()
			if err != nil {
				log.Println(err.Error())
				return
			}

			go h.Core.Email.SendMailNow(viewData, hook.Email.Recipient, "")

		}
	}
}

func (h *Hooker) WriteHistory(db *sql.DB, as *shared.ActionSummary) {
	//, userId uint64, action string, collectionName string, entityId uint64) {

	identity, _ := h.Core.Model.GetIdentityString(db, as.Collection, as.Pk)
	timestamp := time.Now().Unix()

	log.Println("WRITE HISTORY", as.UserId, identity, timestamp, as.Action, as.Collection, as.Pk)

	sql := fmt.Sprintf(`INSERT INTO history 
		(user, identity, timestamp, action, entity, entity_id) VALUES 
		(%d, '%s', %d, '%s', '%s', %d)`,
		as.UserId, identity, timestamp, as.Action, as.Collection, as.Pk)
	//log.Println(sql)

	_, err := db.Exec(sql)
	if err != nil {
		log.Println(err)
		return
	}
}
