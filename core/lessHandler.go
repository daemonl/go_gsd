package core

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

type lessHandler struct {
	filename string
	config   *ServerConfig
}

func (h *lessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("LESS HANDLER")
	w.Header().Add("Content-Type", "text/css")
	c := exec.Command("lessc", h.config.WebRoot+"/"+h.filename)
	c.Stdout = w
	c.Stderr = os.Stderr
	c.Run()
}
