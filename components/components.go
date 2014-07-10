package components

import (
	"database/sql"

	"github.com/daemonl/go_gsd/shared"
)

type Reporter interface {
	GetReportHTMLWriter(name string, pk uint64, session shared.ISession) (shared.IResponse, error)
	Handle(shared.IPathRequest) (shared.IResponse, error)
}

type Mailer interface {
	SendSimple(to string, subject string, body string)
	Send(email *shared.Email) error
	SendResponse(response shared.IResponse, recipient string, notes string) error
}

type PDFer interface {
	ResponseAsPDF(htmlResponse shared.IResponse) (pdfRespnse shared.IResponse)
}

type Runner interface {
	Run(filename string, parameters map[string]interface{}, db *sql.DB) (map[string]interface{}, error)
}

type Hooker interface {
	DoPreHooks(db *sql.DB, as *shared.ActionSummary, session shared.ISession)
	DoPostHooks(db *sql.DB, as *shared.ActionSummary, session shared.ISession)
}

type Xero interface {
	Post(collection string, data interface{}, params ...string) (string, error)
}
