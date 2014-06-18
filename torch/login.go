package torch

import (
	"database/sql"
	"errors"
	"log"
)

type User struct {
	Id             uint64 `json:"id"`
	Username       string `json:"username"`
	password       string
	Access         uint64 `json:"access"`
	SetOnNextLogin bool   `json:"set_on_next_login"`
}

func HandleLogout(requestTorch *Request) {
	requestTorch.Session.User = nil
	requestTorch.NewSession(requestTorch.Session.Store)
	requestTorch.Session.AddFlash("success", "Logged Out")
	requestTorch.Redirect("/")
}

func HandleLogin(requestTorch *Request) {
	username := requestTorch.PostValueString("username")
	password := requestTorch.PostValueString("password")
	doLogin(requestTorch, false, username, password)
}
func ForceLogin(requestTorch *Request, email string) {
	doLogin(requestTorch, true, email, "")
}

func LoadUserById(db *sql.DB, id uint64) (*User, error) {
	rows, err := db.Query(`SELECT id, username, password, access, set_on_next_login FROM staff WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("Could not find a user in the session store")
	}
	user := &User{}
	rows.Scan(&user.Id, &user.Username, &user.password, &user.Access, &user.SetOnNextLogin)
	return user, nil
}

func doLogin(requestTorch *Request, noPassword bool, username string, password string) {

	db := requestTorch.db

	rows, err := db.Query(`SELECT id, username, password, access, set_on_next_login FROM staff WHERE username = ?`, username)
	if err != nil {
		panic(err)
		log.Fatal(err)
		return
	}
	defer rows.Close()

	canHaz := rows.Next()
	if !canHaz {
		if noPassword {
			requestTorch.Session.AddFlash("error", "Database Lookup Error")
		} else {
			requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		}

		log.Print("No can haz user '" + username + "'")

		requestTorch.Redirect("/login")
		return
	}

	user := User{}
	err = rows.Scan(&user.Id, &user.Username, &user.password, &user.Access, &user.SetOnNextLogin)
	if err != nil {
		log.Println("Error on retrieve user from database")
		log.Println(err.Error())
		if noPassword {
			requestTorch.Session.AddFlash("error", "Invalid User Identifier")
		} else {
			requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		}

		requestTorch.Redirect("/login")
		return
	}

	if !noPassword {
		log.Printf("Check Password")
		res, err := user.CheckPassword(password)
		if err != nil {
			log.Println(err.Error())
			requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
			requestTorch.Redirect("/login")
			return
		}
		if !res {
			log.Printf("PASSWORD MISMATCH '%s'\n", password)
			requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
			requestTorch.Redirect("/login")
			return
		}
	}

	target := "/app.html"
	if requestTorch.Session.LoginTarget != nil {
		target = *requestTorch.Session.LoginTarget
	}
	requestTorch.NewSession(requestTorch.Session.Store)
	requestTorch.Session.User = &user
	if user.SetOnNextLogin {
		requestTorch.Redirect("/set_password")
	} else {
		requestTorch.Redirect(target)
	}

	log.Printf("Done Check Password")
}
