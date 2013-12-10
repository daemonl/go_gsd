package core

import (
	"encoding/json"
	"flag"

	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/pdf"

	"io"
	"log"
	"os"
)

type StringSocketMessage struct {
	Message      string
	FunctionName string
	ResponseId   string
}

func (ssm *StringSocketMessage) GetFunctionName() string {
	return ssm.FunctionName
}
func (ssm *StringSocketMessage) GetResponseId() string {
	return ssm.ResponseId
}
func (ssm *StringSocketMessage) PipeMessage(w io.Writer) {
	w.Write([]byte(ssm.Message))
}

var configFilename string
var doSync bool
var forceSync bool

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFilename, "config", wd+"/config.json", "Use Thusly")
	flag.BoolVar(&doSync, "sync", false, "Kick off a db sync instead of serving")
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

func ParseCLI() *ServerConfig {
	flag.Parse()
	log.Println(configFilename)

	var config ServerConfig
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
