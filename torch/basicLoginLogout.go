package torch

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/daemonl/go_gsd/shared"
)

type basicLoginLogout struct {
	db                *sql.DB
	usersTable        string
	ColId             string
	ColUsername       string
	ColPassword       string
	ColSetOnNextLogin string
	ExtraColumns      []string
	LoadUser          func(*sql.Rows) (shared.IUser, error)
}

func GetBasicLoginLogout(db *sql.DB, usersTable string) shared.ILoginLogout {
	lilo := &basicLoginLogout{
		db:                db,
		usersTable:        usersTable,
		ColId:             "id",
		ColUsername:       "username",
		ColPassword:       "password",
		ColSetOnNextLogin: "set_on_next_login",
		ExtraColumns:      []string{"access"},
		LoadUser:          LoadBasicUser,
	}
	return lilo
}

func (lilo *basicLoginLogout) userColString() string {
	allCols := make([]string, len(lilo.ExtraColumns)+4, len(lilo.ExtraColumns)+4)
	allCols[0] = lilo.ColId
	allCols[1] = lilo.ColUsername
	allCols[2] = lilo.ColPassword
	allCols[3] = lilo.ColSetOnNextLogin
	for i, c := range lilo.ExtraColumns {
		allCols[i+4] = c
	}

	return strings.Join(allCols, ", ")
}
func (lilo *basicLoginLogout) HandleLogout(request shared.IRequest) (shared.IResponse, error) {
	//request.Session.shared.IUser = nil
	log.Println("LOGOUT")
	request.ResetSession()
	request.Session().AddFlash("success", "Logged Out")
	return getRedirectResponse("/")
}

func (lilo *basicLoginLogout) HandleLogin(request shared.IRequest) (shared.IResponse, error) {
	username := request.PostValueString("username")
	password := request.PostValueString("password")
	lilo.doLogin(request, false, username, password)
	return nil, nil
}

func (lilo *basicLoginLogout) ForceLogin(request shared.IRequest, email string) {
	lilo.doLogin(request, true, email, "")
}

func (lilo *basicLoginLogout) LoadUserById(id uint64) (shared.IUser, error) {
	rows, err := lilo.db.Query(`SELECT `+lilo.userColString()+` FROM `+lilo.usersTable+` WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("Could not find a user in the session store")
	}
	return lilo.LoadUser(rows)
}

func (lilo *basicLoginLogout) doLogin(request shared.IRequest, noPassword bool, username string, password string) {

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

	rows, err := db.Query(`SELECT `+lilo.userColString()+` FROM `+lilo.usersTable+` WHERE `+lilo.ColUsername+` = ?`, username)
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

	user, err := lilo.LoadUser(rows)
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

	request.Session().SetUser(user)
	if user.SetOnNextLogin() {
		request.Redirect("/set_password")
	} else {
		request.Redirect(target)
	}

	log.Printf("Done Check Password")
}

func (lilo *basicLoginLogout) HandleSetPassword(r shared.IRequest) (shared.IResponse, error) {
	doErr := func(err error) (shared.IResponse, error) {
		log.Println(err)
		r.Session().AddFlash("error", "Something went wrong...")
		return getRedirectResponse("/set_password")

	}
	currentPassword := r.PostValueString("current_password")
	newPassword1 := r.PostValueString("new_password_1")
	newPassword2 := r.PostValueString("new_password_2")

	if newPassword1 != newPassword2 {
		r.Session().AddFlash("error", "Passwords didn't match")
		return getRedirectResponse("/set_password")
	}

	if len(currentPassword) < 1 {
		// Is user exempt?
		//if !r.Session().shared.IUser().SetOnNextLogin {
		//	r.Session.AddFlash("error", "Incorrect current password")
		//	r.Redirect("/set_password")
		//	return
		//}
	} else {
		//Check Current Password
		matches, err := r.Session().User().CheckPassword(currentPassword)
		if err != nil {
			return doErr(err)
		}
		if !matches {
			r.Session().AddFlash("error", "Incorrect current password")
			return getRedirectResponse("/set_password")
		}
	}

	/// Is it secure enough?
	// TODO:... something useful.
	if len(newPassword1) < 5 {
		r.Session().AddFlash("error", "Password must be at least 5 characters long")

		return getRedirectResponse("/set_password")
	}

	hashed := HashPassword(newPassword1)

	db := lilo.db
	_, err := db.Exec(`UPDATE `+lilo.usersTable+` SET `+lilo.ColPassword+` = ?, `+lilo.ColSetOnNextLogin+` = 0 WHERE `+lilo.ColId+` = ?`, hashed, r.Session().UserID())
	if err != nil {
		return doErr(err)
	}
	return getRedirectResponse("/app.html")

}
