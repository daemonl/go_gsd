package quickserve

import (
	"encoding/json"
	"os"
)

type ServerConfig struct {
	DSN             string   `json:"dsn"`
	PublicPatterns  []string `json:"publicPatterns"`
	SessionDumpFile *string  `json:"sessionFile"`
	BindAddress     string   `json:"bind"`
	PublicRoot string `json:"public"`
	TemplateRoot string `json:"templateRoot"`
	TemplateIncludeRoot string `json:"templateInclude"`
}

// LoadConfig reads a json file into an interface.
func LoadConfig(filename string, into interface{}) error {
	configFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer configFile.Close()

	dec := json.NewDecoder(configFile)
	err = dec.Decode(into)
	if err != nil {
		return err
	}
	return nil
}
