package minihandlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type LessHandler struct {
	Filename string
	WebRoot  string
}

func (h *LessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("LESS: Begin")
	w.Header().Add("Content-Type", "text/css")
	c := exec.Command("lessc", h.WebRoot+"/"+h.Filename)
	errBuf := bytes.Buffer{}
	c.Stdout = w
	c.Stderr = &errBuf
	c.Run()
	if errBuf.Len() > 0 {
		log.Println("LESS: Error")
		time.Sleep(time.Second * 1)
		fmt.Fprintf(os.Stderr, "LESS Compile Error (delayed):\n")
		os.Stderr.Write(errBuf.Bytes())
	} else {
		log.Println("LESS: Success")
	}
}
