package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/daemonl/go_gsd/core"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/pdf"
)

var configFilename string
var doSync bool
var forceSync bool

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFilename, "config", wd+"/config.json", "Use Thusly")
	flag.BoolVar(&doSync, "sync", false, "Kick off a db sync instead of serving, Dumps the SQL to stdout unless --force is set")
	flag.BoolVar(&forceSync, "force", false, "Run SQL statements live")
}

func fileNameToObject(filename string, object interface{}) error {
	log.Println("LOAD OBJECT " + filename)
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

func parseCLI() *core.ServerConfig {
	flag.Parse()
	log.Println(configFilename)

	var config core.ServerConfig
	err := fileNameToObject(configFilename, &config)
	if err != nil {
		panic("Could not load config, Aborting: " + err.Error())
	}

	if config.EmailFile != nil {
		var ec email.EmailHandlerConfig
		//ec1 := make(map[string]interface{})
		err := fileNameToObject(*config.EmailFile, &ec)
		if err != nil {
			panic("Could not load config, Aborting: " + err.Error())
		}

		config.EmailConfig = &ec
	}
	if config.PdfFile != nil {

		var ec pdf.PdfHandlerConfig
		//ec1 := make(map[string]interface{})
		err := fileNameToObject(*config.PdfFile, &ec)
		if err != nil {
			panic("Could not load config, Aborting: " + err.Error())
		}

		config.PdfConfig = &ec
	}
	return &config
}

func main() {
	config := parseCLI()

	if doSync {
		err := core.Sync(config, forceSync)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	err := core.Serve(config)
	if err != nil {
		log.Println(err)
	}
}
