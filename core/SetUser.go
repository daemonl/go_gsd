package core

import (
	"log"

	"github.com/daemonl/go_gsd/torch"
)

func SetUser(config *ServerConfig, username string, password string) error {
	core, err := config.GetCore()
	if err != nil {
		return err
	}

	db, err := core.OpenDatabaseConnection(nil)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id FROM staff WHERE username = ?`, username)
	if err != nil {
		return err
	}
	var id uint64 = 0
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Has Next %d\n", id)
	}
	rows.Close()

	hashedPassword := torch.HashPassword(password)

	if id != 0 {

		_, err := db.Exec(`UPDATE staff SET password = ? WHERE id = ?`, hashedPassword, id)
		if err != nil {
			return err
		}

	} else {
		_, err := db.Exec(`INSERT INTO staff (username, password, set_on_next_login, access) VALUES (?, ?, 0, 0)`, username, hashedPassword)

		if err != nil {
			return err
		}

	}
	return nil

}
