package mailhandler

import (
	"github.com/daemonl/go_gsd/components"
	"github.com/daemonl/go_gsd/shared"
)

type MailHandler struct {
	Mailer   components.Mailer
	Reporter components.Reporter
}

func (h *MailHandler) Handle(request shared.IPathRequest) (shared.IResponse, error) {
	reportName := ""
	var id uint64
	recipientRaw := ""
	notes := ""

	err := request.ScanPath(&reportName, &id, &recipientRaw, &notes)
	if err != nil {
		return nil, err
	}
	if notes == "-" {
		notes = ""
	}

	report, err := h.Reporter.GetReportHTMLWriter(reportName, id, request.Session())
	if err != nil {
		return nil, err
	}

	err = h.Mailer.SendMailFromResponse(report, recipientRaw, notes)

	if err != nil {
		return nil, err
	}

	return shared.QuickStringResponse("Email sent to " + recipientRaw), nil
}
