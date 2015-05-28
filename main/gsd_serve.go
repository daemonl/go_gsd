package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/daemonl/go_gsd/core"
	"github.com/daemonl/go_gsd/reporter"
)

var configFilename string
var doSync bool
var forceSync bool
var setUser string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFilename, "config", wd+"/config.json", "Use Thusly")
	flag.BoolVar(&doSync, "sync", false, "Kick off a db sync instead of serving, Dumps the SQL to stdout unless --force is set")
	flag.BoolVar(&forceSync, "force", false, "Run SQL statements live")
}

func parseCLI() (*core.ServerConfig, error) {
	flag.Parse()
	log.Printf("Load config from %s\n", configFilename)

	var config core.ServerConfig
	err := core.FileNameToObject(configFilename, &config)
	if err != nil {
		return nil, err

	}

	return &config, nil
}

func main() {
	config, err := parseCLI()
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}

	if doSync {
		err := core.Sync(config, forceSync)
		if err != nil {
			log.Println(err)
			return
		}
	}
	if len(setUser) > 0 {
		parts := strings.Split(setUser, ":")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "setuser must be in the form username:password")
			os.Exit(1)
		}
		core.SetUser(config, parts[0], parts[1])
	}

	if doSync {
		return
	}

	err = core.Serve(config)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}
}
