package minihandlers

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

type LessHandler struct {
	Filename string
	WebRoot  string
}

func (h *LessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("LESS HANDLER")
	w.Header().Add("Content-Type", "text/css")
	c := exec.Command("lessc", h.WebRoot+"/"+h.Filename)
	c.Stdout = w
	c.Stderr = os.Stderr
	c.Run()
}
