package pdf

import (
	"bytes"
	"github.com/daemonl/go_gsd/shared"
	"github.com/daemonl/go_gsd/view"

	"fmt"
)

type PDFHandler struct {
	HandlerConfig  *PDFHandlerConfig
	TemplateWriter *view.TemplateWriter
	Binary         string
}

type PDFHandlerConfig struct {
	Templates map[string]view.TemplateConfig `json:"templates"`
}

func GetPDFHandler(binary string, handlerConfig *PDFHandlerConfig, templateWriter *view.TemplateWriter) (*PDFHandler, error) {
	eh := PDFHandler{
		HandlerConfig:  handlerConfig,
		TemplateWriter: templateWriter,
		Binary:         binary,
	}
	return &eh, nil
}

func (h *PDFHandler) GetReport(request shared.IPathRequest) (*view.Report, error) {
	reportName := ""
	var id uint64

	err := request.ScanPath(&reportName, &id)
	if err != nil {
		return nil, err
	}

	reportConfig, ok := h.HandlerConfig.Templates[reportName]
	if !ok {
		return nil, fmt.Errorf("Report %s not found", reportName)
	}

	r := &view.Report{
		Session: request.Session(),
		Config:  &reportConfig,
		RootID:  id,
	}
	return r, nil
}

func (h *PDFHandler) Preview(request shared.IPathRequest) (shared.IResponse, error) {

	report, err := h.GetReport(request)
	if err != nil {
		return nil, err
	}

	viewData, err := report.PrepareData()
	if err != nil {
		return nil, err
	}

	return viewData, nil
}

func (h *PDFHandler) GetPDF(request shared.IPathRequest) (shared.IResponse, error) {

	report, err := h.GetReport(request)
	if err != nil {
		return nil, err
	}

	viewData, err := report.PrepareData()
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	err = viewData.WriteTo(w)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(w.Bytes())

	pdfResponse := &PDFResponse{
		hTMLIn: r,
		binary: h.Binary,
	}
	return pdfResponse, nil
}
