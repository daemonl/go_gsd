package core

import (
	"encoding/json"
	"fmt"

	"github.com/daemonl/databath/sync"
)

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
	mig, err := sync.BuildMigration(db, core.GetModel())
	if err != nil {
		return err
	}

	e, err := json.Marshal(mig)
	if err != nil {
		return err
	}
	fmt.Println(string(e))

	if force {
		err := mig.Run(db)
		if err != nil {
			return err
		}
	}
	return nil
}
