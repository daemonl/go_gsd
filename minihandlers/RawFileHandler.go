package minihandlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type RawFileHandler struct {
	Filename string
	WebRoot  string
}

func (h *RawFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open(h.WebRoot + "/" + h.Filename)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
