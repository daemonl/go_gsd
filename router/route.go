package router

import (
	"regexp"
)

type route struct {
	re      *regexp.Regexp
	format  string
	handler Handler //func(Request) (Response, error)
	methods []string
}
