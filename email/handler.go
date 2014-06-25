package email

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/router"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"log"
	"strings"
)

type EmailHandler struct {
	SmtpConfig     *SmtpConfig
	Sender         *Sender
	HandlerConfig  *EmailHandlerConfig
	Model          *databath.Model
	TemplateWriter *view.TemplateWriter
}

type EmailHook struct {
	Collection   string `json:"collection"`
	TriggerField string `json:"triggerField"`
	Recipient    string `json:"recipient"`
	Template     string `json:"template"`
}
type EmailHandlerConfig struct {
	From        string                         `json:"from"`
	TemplateDir string                         `json:"templateDir"`
	Templates   map[string]view.TemplateConfig `json:"templates"`
	Hooks       []EmailHook                    `json:"hooks"`
}

func GetEmailHandler(smtpConfig *SmtpConfig, handlerConfig *EmailHandlerConfig, templateWriter *view.TemplateWriter) (*EmailHandler, error) {
	eh := EmailHandler{
		SmtpConfig:     smtpConfig,
		HandlerConfig:  handlerConfig,
		TemplateWriter: templateWriter,
	}
	sender := Sender{
		Config: smtpConfig,
	}
	eh.Sender = &sender
	return &eh, nil
}

func (h *EmailHandler) GetReport(reportName string, rootID uint64, session torch.Session) (*view.Report, error) {

	reportConfig, ok := h.HandlerConfig.Templates[reportName]
	if !ok {
		return nil, fmt.Errorf("Report %s not found", reportName)
	}

	r := &view.Report{
		Session: session,
		Config:  &reportConfig,
		RootID:  rootID,
	}

	return r, nil
}

func (h *EmailHandler) Preview(request router.Request) (router.Response, error) {

	reportName := ""
	var id uint64

	err := request.ScanPath(&reportName, &id)
	if err != nil {
		return nil, err
	}

	report, err := h.GetReport(reportName, id, request.Session())
	if err != nil {
		return nil, err
	}

	viewData, err := report.PrepareData()
	if err != nil {
		return nil, err
	}

	return viewData, nil
}

func (h *EmailHandler) Send(request router.Request) (router.Response, error) {

	emailName := ""
	var id uint64
	recipientRaw := ""
	notes := ""

	err := request.ScanPath(&emailName, &id, &recipientRaw, &notes)
	if err != nil {
		return nil, err
	}

	report, err := h.GetReport(emailName, id, request.Session())
	if err != nil {
		return nil, err
	}

	data, err := report.PrepareData()
	if err != nil {
		return nil, err
	}

	recipients := strings.Split(recipientRaw, ";")
	for _, recipient := range recipients {
		err := h.SendMailNow(data, strings.TrimSpace(recipient), notes)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func dropLine(in *string) (string, error) {
	log.Println("Drop Line")

	parts := strings.SplitN(*in, "\n", 2)
	if len(parts) != 2 {
		return "", errors.New("Could not extract line")
	}
	*in = parts[1]

	return parts[0], nil
}

func (h *EmailHandler) SendMailNow(data *view.ViewData, recipient string, notes string) error {

	w := &bytes.Buffer{}

	err := data.WriteTo(w)
	if err != nil {
		return err
	}
	html := w.String()
	subject, err := dropLine(&html)
	if err != nil {
		return err
	}

	if recipient == "#inline" {
		recipient, err = dropLine(&html)
		if err != nil {
			return err
		}
	}

	notes = strings.Replace(notes, "\n", "<br/>", -1)
	html = strings.Replace(html, "--- NOTES HERE ---", notes, 1)

	email := Email{
		Sender:    h.HandlerConfig.From,
		Recipient: recipient,
		Subject:   subject,
		Html:      html,
	}

	err = h.Sender.Send(&email)
	if err != nil {
		return err
	}
	return nil
}
