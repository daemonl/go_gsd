package torch

import (
	"log"
)

func HandleSetPassword(r *Request) {
	currentPassword := r.PostValueString("current_password")
	newPassword1 := r.PostValueString("new_password_1")
	newPassword2 := r.PostValueString("new_password_2")

	if newPassword1 != newPassword2 {
		r.Session.AddFlash("error", "Passwords didn't match")
		r.Redirect("/set_password")
		return
	}

	if len(currentPassword) < 1 {
		// Is user exempt?
		if !r.Session.User.SetOnNextLogin {
			r.Session.AddFlash("error", "Incorrect current password")
			r.Redirect("/set_password")
			return
		}
	} else {
		//Check Current Password
		matches, err := CheckPassword(r.Session.User.Password, currentPassword)
		if err != nil {
			log.Println(err)
			r.Session.AddFlash("error", "Something went wrong...")
			r.Redirect("/set_password")
			return
		}
		if !matches {
			r.Session.AddFlash("error", "Incorrect current password")
			r.Redirect("/set_password")
			return
		}
	}

	/// Is it secure enough?
	if len(newPassword1) < 5 {
		r.Session.AddFlash("error", "Password must be at least 5 characters long")
		r.Redirect("/set_password")
		return
	}

	hashed := HashPassword(newPassword1)

	db := r.DbConn.GetDB()
	_, err := db.Exec(`UPDATE staff SET password = ?, set_on_next_login = 0 WHERE id = ?`, hashed, r.Session.User.Id)
	if err != nil {
		log.Println(err)
		r.Session.AddFlash("error", "Something went wrong...")
		r.Redirect("/set_password")
		return
	}
	r.Redirect("/app.html")
}
