package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type rawFileHandler struct {
	filename string
	config   *ServerConfig
}

func (h *rawFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open(h.config.WebRoot + "/" + h.filename)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
