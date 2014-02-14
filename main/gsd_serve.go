package main

import (
	"github.com/daemonl/go_gsd/core"
)

func main() {
	config := core.ParseCLI()
	core.Serve(config)
}
