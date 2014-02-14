package core

import (
	"github.com/daemonl/go_gsd/torch"
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
func (h *FallthroughHandler) Handle(requestTorch *torch.Request) {
	_, r := requestTorch.GetRaw()
	if r.URL.Path == "/" {
		requestTorch.Redirect("/app.html")
	}
	h.fsHandler.ServeHTTP(requestTorch.GetRaw())
}
