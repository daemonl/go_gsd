package minihandlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/daemonl/go_gsd/shared"
)

type RawFileHandler struct {
	Filename string
	WebRoot  string
}

func (h *RawFileHandler) Handle(req shared.IRequest) (shared.IResponse, error) {
	w, _ := req.GetRaw()
	file, err := os.Open(h.WebRoot + "/" + h.Filename)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil
	}
	defer file.Close()
	io.Copy(w, file)
	return nil, nil
}
