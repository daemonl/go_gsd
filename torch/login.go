package torch

import (
	"log"
)

type User struct {
	Id             uint64 `json:"id"`
	Username       string `json:"username"`
	Password       string `json:""`
	Access         string `json:"access"`
	SetOnNextLogin bool   `json:""`
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

	db := requestTorch.DbConn.GetDB()

	rows, err := db.Query(`SELECT id, username, password, access, set_on_next_login FROM staff WHERE username = ?`, username)
	if err != nil {
		panic(err)
		log.Fatal(err)
		return
	}

	canHaz := rows.Next()
	if !canHaz {
		log.Print("No can haz user '" + username + "'")
		requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		requestTorch.Redirect("/login")
		return
	}

	user := User{}
	err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Access, &user.SetOnNextLogin)
	if err != nil {
		log.Println("Error on retrieve user from database")
		log.Println(err.Error())
		requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		requestTorch.Redirect("/login")
		return
	}

	log.Printf("Check Password")
	res, err := CheckPassword(user.Password, password)
	if err != nil {
		log.Println(err.Error())
		requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		requestTorch.Redirect("/login")
		return
	}
	if res {
		requestTorch.NewSession(requestTorch.Session.Store)
		requestTorch.Session.User = &user
		if user.SetOnNextLogin {
			requestTorch.Redirect("/set_password")
		} else {
			requestTorch.Redirect("/app.html")
		}

	} else {
		requestTorch.Session.AddFlash("error", "The presented credentials were incorrect. Please try again.")
		requestTorch.Redirect("/login")
	}
	log.Printf("Done Check Password")
}
