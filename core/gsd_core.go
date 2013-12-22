package core

import (
	"github.com/daemonl/go_gsd/email"

	"github.com/daemonl/go_lib/databath"
)

type GSDCore struct {
	Bath   *databath.Bath
	Model  *databath.Model
	Email  *email.EmailHandler
	Hooker *Hooker
	Config *ServerConfig
}
