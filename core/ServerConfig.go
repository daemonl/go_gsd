package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/csv"
	"github.com/daemonl/go_gsd/dynamic"
	"github.com/daemonl/go_gsd/hooker"
	"github.com/daemonl/go_gsd/mailer"
	"github.com/daemonl/go_gsd/mailhandler"
	"github.com/daemonl/go_gsd/pdfer"
	"github.com/daemonl/go_gsd/pdfhandler"
	"github.com/daemonl/go_gsd/reporter"
	"github.com/daemonl/go_gsd/view"

	"github.com/daemonl/go_gsd/google_auth"
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
	S3 struct {
		BucketName string `json:"bucket"`
	} `json:"s3"`
	ModelFile           string   `json:"modelFile"`
	TemplateRoot        string   `json:"templateRoot"`
	WebRoot             string   `json:"webRoot"`
	BindAddress         string   `json:"bindAddress"`
	PublicPatternsRaw   []string `json:"publicPatterns"`
	UploadDirectory     string   `json:"uploadDirectory"`
	TemplateIncludeRoot string   `json:"templateIncludeRoot"`
	ScriptDirectory     string   `json:"scriptDirectory"`
	Timezone            *string  `json:"timezone"`

	ReportFile *string `json:"reportFile"`
	Reports    map[string]reporter.ReportConfig

	SmtpConfig      *mailer.SmtpConfig `json:"smtpConfig"`
	SessionDumpFile *string            `json:"sessionDumpFile"`

	PDFBinary *string `json:"pdfBinary"`

	GoogleAuthConfig *google_auth.GoogleAuth `json:"googleAuth"`
}

func FileNameToObject(filename string, object interface{}) error {
	filename = os.ExpandEnv(filename)
	fmt.Println(filename)
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

func Expand(in *string) {
	if in == nil {
		return
	}
	*in = os.ExpandEnv(*in)
}

func (config *ServerConfig) GetCore() (core *GSDCore, err error) {

	Expand(&config.Xero.PrivateKeyFile)
	Expand(&config.ModelFile)
	Expand(&config.TemplateRoot)
	Expand(&config.WebRoot)
	Expand(&config.UploadDirectory)
	Expand(&config.TemplateIncludeRoot)
	Expand(&config.ScriptDirectory)
	Expand(config.ReportFile)
	Expand(config.SessionDumpFile)

	timezoneName := "Australia/Melbourne"
	if config.Timezone != nil {
		timezoneName = *config.Timezone
	}
	loc, err := time.LoadLocation(timezoneName) //"Australia/Melbourne")
	if err != nil {
		return nil, err
	}
	time.Local = loc

	core = &GSDCore{}

	if config.ReportFile != nil {
		log.Printf("Load email config from %s\n", *config.ReportFile)
		var rc map[string]reporter.ReportConfig

		err := FileNameToObject(*config.ReportFile, &rc)
		if err != nil {
			return nil, err
		}
		config.Reports = rc
	}

	//////////////////////
	// Template Manager //
	log.Println("Load Template Manager")
	templateManager := view.GetTemplateManager(string(config.TemplateRoot), string(config.TemplateIncludeRoot))
	core.TemplateManager = templateManager

	model, err := databath.ReadModelFromFile(string(config.ModelFile))
	if err != nil {
		return nil, fmt.Errorf("Error reading model file: %s", err.Error())
	}
	core.Model = model

	////////////
	// Mailer //
	log.Println("Load Mailer")
	mailer := &mailer.Mailer{
		Config: config.SmtpConfig,
	}
	core.Mailer = mailer

	/////////////
	// Runner //
	log.Println("Load Script Runner")
	runner := &dynamic.DynamicRunner{
		BaseDirectory: string(config.ScriptDirectory),
		Mailer:        mailer,
		Xero:          core.Xero,
	}
	core.Runner = runner

	//////////////
	// Reporter //
	log.Println("Load Reporter")
	reporter := &reporter.Reporter{
		ViewManager: templateManager,
		Runner:      runner,
		Model:       model,
		Reports:     config.Reports,
	}
	core.Reporter = reporter

	///////////
	// PDFer //
	log.Println("Load PDFer")
	pdfer := &pdfer.PDFer{
		Binary: *config.PDFBinary,
	}
	core.PDFer = pdfer

	////////////////
	// PDFHandler //
	log.Println("Load PDF Handler")
	pdfHandler := &pdfhandler.PDFHandler{
		Reporter: reporter,
		PDFer:    pdfer,
	}
	core.PDFHandler = pdfHandler

	//////////////////
	// MailHandler //
	log.Println("Load Mail Handler")
	mailHandler := &mailhandler.MailHandler{
		Mailer:   mailer,
		Reporter: reporter,
	}
	core.MailHandler = mailHandler

	/////////
	// CSV //
	log.Println("Load CSV Handler")
	csv := &csv.CSVHandler{
		Model: model,
	}
	core.CSVHandler = csv

	////////////
	// Hooker //
	log.Println("Load Hooker")
	hooker := &hooker.Hooker{
		Model:    model,
		Runner:   runner,
		Reporter: reporter,
		Mailer:   mailer,
	}
	core.Hooker = hooker

	core.Config = config

	log.Println("Connect to Database")
	db, err := sql.Open(core.Config.Database.Driver, core.Config.Database.DataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Database ping failed: %s", err.Error())
	}

	core.DB = db

	log.Println("End Config Loading")

	return core, err

}
