package pdfhandler

import (
	"github.com/daemonl/go_gsd/components"
	"github.com/daemonl/go_gsd/shared"
)

type PDFHandler struct {
	Reporter components.Reporter
	PDFer    components.PDFer
}

func (h *PDFHandler) Handle(request shared.IPathRequest) (shared.IResponse, error) {
	reportName := ""
	var id uint64

	err := request.ScanPath(&reportName, &id)
	if err != nil {
		return nil, err
	}

	report, err := h.Reporter.GetReportHTMLWriter(reportName, id, request.Session())
	if err != nil {
		return nil, err
	}

	pdfResponse := h.PDFer.ResponseAsPDF(report)
	return pdfResponse, nil
}
