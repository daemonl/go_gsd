package components

import (
	"database/sql"

	"github.com/daemonl/go_gsd/shared"
)

type Reporter interface {
	GetReportHTMLWriter(name string, pk uint64, session shared.ISession) (shared.IResponse, error)
	HandleReportRequest(shared.IPathRequest) (shared.IResponse, error)
}

type Mailer interface {
	SendMailSimple(to string, subject string, body string)
	SendMail(email *shared.Email) error
	SendMailFromResponse(response shared.IResponse, recipient string, notes string) error
}

type PDFer interface {
	ResponseAsPDF(htmlResponse shared.IResponse) (pdfRespnse shared.IResponse)
}

type Runner interface {
	RunScript(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
}

type Hooker interface {
	DoPreHooks(hc *HookContext)
	DoPostHooks(hc *HookContext)
}

type Xero interface {
	XeroPost(collection string, data interface{}, params ...string) (string, error)
}

type Core interface {
	Reporter
	Mailer
	PDFer
	Runner
	Hooker
	Xero
}

type HookContext struct {
	DB            *sql.DB
	ActionSummary *shared.ActionSummary
	Session       shared.ISession
	Core          Core
}
