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
var devMode bool
var setUser string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFilename, "config", wd+"/config.json", "Use Thusly")
	flag.BoolVar(&doSync, "sync", false, "Kick off a db sync instead of serving, Dumps the SQL to stdout unless --force is set")
	flag.BoolVar(&forceSync, "force", false, "Run SQL statements live")
	flag.BoolVar(&devMode, "dev", false, "Use app_dev.html and compile less live")
	flag.StringVar(&setUser, "setuser", "", "specify a user for development username:password, will create or update")
}

func fileNameToObject(filename string, object interface{}) error {
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(object)

	if err != nil {
		return err
	}

	return nil
}

func parseCLI() (*core.ServerConfig, error) {
	flag.Parse()
	log.Printf("Load config from %s\n", configFilename)

	var config core.ServerConfig
	err := fileNameToObject(configFilename, &config)
	if err != nil {
		return nil, err

	}

	config.DevMode = devMode

	if config.ReportFile != nil {
		log.Printf("Load email config from %s\n", *config.ReportFile)
		var rc map[string]reporter.ReportConfig

		err := fileNameToObject(*config.ReportFile, &rc)
		if err != nil {
			return nil, err
		}
		config.Reports = rc
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
