package core

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFilename, "config", wd+"/config.json", "Use Thusly")
}

func ParseCLI() *ServerConfig {
	flag.Parse()
	fmt.Println(configFilename)
	configFile, err := os.Open(configFilename)
	if err != nil {
		panic("Could Not Read Config File: " + err.Error())
	}

	var config ServerConfig

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic("Could Not Decode Config File: " + err.Error())
	}
	configFile.Close()
	return &config
}
