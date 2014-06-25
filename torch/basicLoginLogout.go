package torch

import (
	"database/sql"
	"errors"
	"log"
)

type basicLoginLogout struct {
	db *sql.DB
}

func GetBasicLoginLogout(db *sql.DB) LoginLogout {
	lilo := &basicLoginLogout{
		db: db,
	}
	return lilo
}

func (lilo *basicLoginLogout) HandleLogout(request Request) {
	//request.Session.User = nil
	request.ResetSession()
	request.Session().AddFlash("success", "Logged Out")
	request.Redirect("/")
}

func (lilo *basicLoginLogout) HandleLogin(request Request) {
	username := request.PostValueString("username")
	password := request.PostValueString("password")
	lilo.doLogin(request, false, username, password)
}

func (lilo *basicLoginLogout) ForceLogin(request Request, email string) {
	lilo.doLogin(request, true, email, "")
}

func (lilo *basicLoginLogout) LoadUserById(id uint64) (User, error) {
	rows, err := lilo.db.Query(`SELECT id, username, password, access, set_on_next_login FROM staff WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("Could not find a user in the session store")
	}
	user := &basicUser{}
	rows.Scan(&user.id, &user.username, &user.password, &user.access, &user.setOnNextLogin)
	return user, nil
}

func (lilo *basicLoginLogout) doLogin(request Request, noPassword bool, username string, password string) {

	doError := func(verboseMessage string, err error) {
		log.Printf("Issue loggin in (not error): %s, U:%s", verboseMessage, username)

		if err != nil {
			log.Printf("Error loading user '%s' from database: %s\n", username, err.Error())
		}
		if noPassword {
			request.Session().AddFlash("error", verboseMessage)
		} else {
			request.Session().AddFlash("error", "The presented credentials were not matched. Please try again.")
		}
		request.Redirect("/login")
	}

	db := lilo.db

	rows, err := db.Query(`SELECT id, username, password, access, set_on_next_login FROM staff WHERE username = ?`, username)
	if err != nil {
		panic(err)
		log.Fatal(err)
		return
	}

	defer rows.Close()

	canHaz := rows.Next()
	if !canHaz {
		doError("Database lookup error", nil)
		return
	}

	user := basicUser{}
	err = rows.Scan(&user.id, &user.username, &user.password, &user.access, &user.setOnNextLogin)
	if err != nil {
		doError("Invalid user identifier", err)
		return
	}

	if !noPassword {
		log.Printf("Check Password")
		res, err := user.CheckPassword(password)
		if err != nil {
			doError("", err)
			return
		}
		if !res {
			doError("", err)
			return
		}
	}

	target := "/app.html"
	//if request.Session.LoginTarget != nil {
	//		target = *request.Session.LoginTarget
	//	}

	request.ResetSession()

	request.Session().SetUser(&user)
	if user.setOnNextLogin {
		request.Redirect("/set_password")
	} else {
		request.Redirect(target)
	}

	log.Printf("Done Check Password")
}

func (lilo *basicLoginLogout) HandleSetPassword(r Request) {
	doErr := func(err error) {
		log.Println(err)
		r.Session().AddFlash("error", "Something went wrong...")
		r.Redirect("/set_password")
	}
	currentPassword := r.PostValueString("current_password")
	newPassword1 := r.PostValueString("new_password_1")
	newPassword2 := r.PostValueString("new_password_2")

	if newPassword1 != newPassword2 {
		r.Session().AddFlash("error", "Passwords didn't match")
		r.Redirect("/set_password")
		return
	}

	if len(currentPassword) < 1 {
		// Is user exempt?
		//if !r.Session().User().SetOnNextLogin {
		//	r.Session.AddFlash("error", "Incorrect current password")
		//	r.Redirect("/set_password")
		//	return
		//}
	} else {
		//Check Current Password
		matches, err := r.Session().User().CheckPassword(currentPassword)
		if err != nil {
			doErr(err)
			return
		}
		if !matches {
			r.Session().AddFlash("error", "Incorrect current password")
			r.Redirect("/set_password")
			return
		}
	}

	/// Is it secure enough?
	// TODO:... something useful.
	if len(newPassword1) < 5 {
		r.Session().AddFlash("error", "Password must be at least 5 characters long")
		r.Redirect("/set_password")
		return
	}

	hashed := HashPassword(newPassword1)

	db := lilo.db
	_, err := db.Exec(`UPDATE staff SET password = ?, set_on_next_login = 0 WHERE id = ?`, hashed, r.Session().UserID())
	if err != nil {
		doErr(err)
		return
	}
	r.Redirect("/app.html")
}
