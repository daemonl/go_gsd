package core

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/csv"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/dynamic_xero"
	"github.com/daemonl/go_gsd/hooker"
	"github.com/daemonl/go_gsd/mailer"
	"github.com/daemonl/go_gsd/mailhandler"
	"github.com/daemonl/go_gsd/pdfer"
	"github.com/daemonl/go_gsd/pdfhandler"
	"github.com/daemonl/go_gsd/reporter"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_xero"

	"github.com/daemonl/go_lib/google_auth"
)

type ServerConfig struct {
	Database struct {
		Driver         string `json:"driver"`
		DataSourceName string `json:"dsn"`
		PoolSize       int    `json:"poolSize"`
	} `json:"database"`
	Xero struct {
		AppKey         string `json:"appKey"`
		PrivateKeyFile string `json:"privateKeyFile"`
	} `json:"xero"`
	ModelFile           string   `json:"modelFile"`
	TemplateRoot        string   `json:"templateRoot"`
	WebRoot             string   `json:"webRoot"`
	BindAddress         string   `json:"bindAddress"`
	PublicPatternsRaw   []string `json:"publicPatterns"`
	UploadDirectory     string   `json:"uploadDirectory"`
	TemplateIncludeRoot string   `json:"templateIncludeRoot"`
	ScriptDirectory     string   `json:"scriptDirectory"`
	DevMode             bool

	ReportFile *string `json:"reportFile"`
	Reports    map[string]reporter.ReportConfig

	SmtpConfig      *mailer.SmtpConfig `json:"smtpConfig"`
	SessionDumpFile *string            `json:"sessionDumpFile"`

	PDFBinary *string `json:"pdfBinary"`

	OAuthConfig *google_auth.OAuthConfig `json:"oauth"`
}

func (config *ServerConfig) GetCore() (core *GSDCore, err error) {

	core = &GSDCore{}

	//////////////////////
	// Template Manager //
	templateManager := view.GetTemplateManager(config.TemplateRoot, config.TemplateIncludeRoot)
	core.TemplateManager = templateManager

	model, err := databath.ReadModelFromFile(config.ModelFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading model file: %s", err.Error())
	}
	core.Model = model

	////////////
	// Mailer //
	mailer := &mailer.Mailer{
		Config: config.SmtpConfig,
	}
	core.Mailer = mailer

	//////////
	// Xero //
	if len(config.Xero.AppKey) > 0 {
		log.Println("LOAD XERO")
		x, err := xero.GetXeroPrivateCore(config.Xero.PrivateKeyFile, config.Xero.AppKey)
		if err != nil {
			return nil, err
		}
		dx := &dynamic_xero.DynamicXero{
			Xero: x,
		}
		core.Xero = dx
	}

	/////////////
	// Runner //
	runner := &dynamic.DynamicRunner{
		BaseDirectory: config.ScriptDirectory,
		Mailer:        mailer,
		Xero:          core.Xero,
	}
	core.Runner = runner

	//////////////
	// Reporter //
	reporter := &reporter.Reporter{
		ViewManager: templateManager,
		Runner:      runner,
		Model:       model,
		Reports:     config.Reports,
	}
	core.Reporter = reporter

	///////////
	// PDFer //
	pdfer := &pdfer.PDFer{
		Binary: *config.PDFBinary,
	}
	core.PDFer = pdfer

	////////////////
	// PDFHandler //
	pdfHandler := &pdfhandler.PDFHandler{
		Reporter: reporter,
		PDFer:    pdfer,
	}
	core.PDFHandler = pdfHandler

	//////////////////
	// MailHandler //
	mailHandler := &mailhandler.MailHandler{
		Mailer:   mailer,
		Reporter: reporter,
	}
	core.MailHandler = mailHandler

	/////////
	// CSV //
	csv := &csv.CSVHandler{
		Model: model,
	}
	core.CSVHandler = csv

	////////////
	// Hooker //
	hooker := &hooker.Hooker{
		Model:    model,
		Runner:   runner,
		Reporter: reporter,
		Mailer:   mailer,
	}
	core.Hooker = hooker

	core.Config = config

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
