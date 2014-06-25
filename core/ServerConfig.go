package core

import (
	"database/sql"
	"fmt"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/email"
	"github.com/daemonl/go_gsd/pdf"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_lib/google_auth"
)

type ServerConfig struct {
	Database struct {
		Driver         string `json:"driver"`
		DataSourceName string `json:"dsn"`
		PoolSize       int    `json:"poolSize"`
	} `json:"database"`
	ModelFile           string   `json:"modelFile"`
	TemplateRoot        string   `json:"templateRoot"`
	WebRoot             string   `json:"webRoot"`
	BindAddress         string   `json:"bindAddress"`
	PublicPatternsRaw   []string `json:"publicPatterns"`
	UploadDirectory     string   `json:"uploadDirectory"`
	TemplateIncludeRoot string   `json:"templateIncludeRoot"`
	ScriptDirectory     string   `json:"scriptDirectory"`
	DevMode             bool

	EmailConfig     *email.EmailHandlerConfig
	EmailFile       *string           `json:"emailFile"`
	SmtpConfig      *email.SmtpConfig `json:"smtpConfig"`
	SessionDumpFile *string           `json:"sessionDumpFile"`

	PdfConfig *pdf.PdfHandlerConfig
	PdfFile   *string `json:"pdfFile"`
	PdfBinary *string `json:"pdfBinary"`

	OAuthConfig *google_auth.OAuthConfig `json:"oauth"`

	ViewManager *view.ViewManager
}

func (config *ServerConfig) ReloadHandle(request torch.Request) {
	err := config.ViewManager.Reload()
	if err != nil {
		request.Writef("Error Loading Views: %s", err.Error())
		return
	}
	request.Writef("Loaded Views")
}

func (config *ServerConfig) GetCore() (core *GSDCore, err error) {

	model, err := databath.ReadModelFromFile(config.ModelFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading model file: %s", err.Error())
	}

	core = &GSDCore{
		Model:  model,
		Config: config,
	}

	core.Runner = &dynamic.DynamicRunner{
		BaseDirectory: core.Config.ScriptDirectory, // "/home/daemonl/schkit/impl/pov/script/",
		SendMail:      core.SendMail,
	}

	core.Hooker = &Hooker{
		Core: core,
	}

	db, err := sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Database ping failed: %s", err.Error())
	}

	core.DB = db

	return core, err

}
