package core

import (
	"github.com/daemonl/go_gsd/shared"
	"net/http"
)

type FallthroughHandler struct {
	config    *ServerConfig
	fsHandler http.Handler
}

func GetFallthroughHandler(config *ServerConfig) *FallthroughHandler {
	fsHandler := http.FileServer(http.Dir(config.WebRoot))
	h := FallthroughHandler{
		config:    config,
		fsHandler: fsHandler,
	}
	return &h

}
func (h *FallthroughHandler) Handle(request shared.IRequest) {
	_, r := request.GetRaw()
	if r.URL.Path == "/" {
		request.Redirect("/app.html")
	}
	h.fsHandler.ServeHTTP(request.GetRaw())
}
