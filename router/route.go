package router

import (
	"github.com/daemonl/go_gsd/shared"
	"regexp"
)

type route struct {
	re      *regexp.Regexp
	format  string
	handler shared.IHandler //func(Request) (Response, error)
	methods []string
}
