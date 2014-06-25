package core

import (
	"log"
	"net/http"
	"strings"
)

func (core *GSDCore) runScript(w http.ResponseWriter, r *http.Request) {
	// This should only run for requests internal to the system, for cronjobs and scripts.
	// Nginx sets x-real-ip on all forwarded requests.
	// This needs a security review.

	real_ip := r.Header.Get("X-Real-IP")
	if len(real_ip) > 0 {
		w.Write([]byte("NOT ALLOWED"))
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Printf("/script/ called with %s\n", parts)
		return
	}
	script := re_unsafe.ReplaceAllString(parts[2], "_")

	log.Printf("RUN SCRIPT %s\n", script)

	db, err := core.UsersDatabase()
	if err != nil {
		log.Println(err)
		w.Write([]byte("ERROR"))
		return
	}
	_, err = core.Runner.Run(script+".js", map[string]interface{}{"path": parts[2:]}, db)
	if err != nil {
		log.Println(err)
		w.Write([]byte("ERROR"))
		return
	}
	w.Write([]byte("OK"))
}
