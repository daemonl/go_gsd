package core

import (
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_lib/databath"
	"log"
	"time"
)

//type Hook struct {
//	Collection string                 `json:"collection"`
//	When       HookWhen               `json:"when"`
//	Set        map[string]interface{} `json:"set"`
//	Email      *HookEmail              `json:"email"`
//}
//type HookWhen struct {
//	Field string `json:"field"`
//	What  string `json:"what"`
//}
//type HookEmail struct {
//	Recipient string `json:"recipient"`
//	Template  string `json:"template"`
//}

type Hooker struct {
	Core *GSDCore
}

type ActionSummary struct {
	User       *torch.User
	Action     string
	Collection *databath.Collection
	Pk         uint64
	Fields     map[string]interface{}
}

func (h *Hooker) DoPreHooks(as *ActionSummary) {
	for _, hook := range as.Collection.Hooks {
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
					v = as.User.Id
				}
			}
			as.Fields[k] = v
		}

	}
}
func (h *Hooker) DoPostHooks(as *ActionSummary) {
	go h.WriteHistory(as)
	for _, hook := range as.Collection.Hooks {
		log.Println("Q$#@ERWSG%VDY")
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
			results, err := hook.CustomAction.Run(h.Core.Bath, p)
			if err != nil {
				log.Println(err.Error())
				return
			}
			log.Println(results)

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
			go h.Core.Email.SendMailNow(hook.Email.Template, as.Pk, hook.Email.Recipient, "", nil)

		}
	}
}

func (h *Hooker) WriteHistory(as *ActionSummary) {
	//, userId uint64, action string, collectionName string, entityId uint64) {
	identity, _ := h.Core.Model.GetIdentityString(h.Core.Bath, as.Collection.TableName, as.Pk)
	timestamp := time.Now().Unix()

	log.Println("WRITE HISTORY", as.User.Id, identity, timestamp, as.Action, as.Collection.TableName, as.Pk)

	sql := fmt.Sprintf(`INSERT INTO history 
		(user, identity, timestamp, action, entity, entity_id) VALUES 
		(%d, '%s', %d, '%s', '%s', %d)`,
		as.User.Id, identity, timestamp, as.Action, as.Collection.TableName, as.Pk)
	//log.Println(sql)
	c := h.Core.Bath.GetConnection()
	db := c.GetDB()
	defer c.Release()
	_, err := db.Exec(sql)
	if err != nil {
		log.Println(err)
		return
	}
}
