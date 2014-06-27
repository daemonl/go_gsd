package core

import (
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
	sync.SyncDb(db, core.GetModel(), force)
	return nil
}
