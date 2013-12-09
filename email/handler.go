package email

import (
	"bytes"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/go_lib/databath"

	"log"
	"strings"
)

type EmailHandler struct {
	SmtpConfig     *SmtpConfig
	Sender         *Sender
	HandlerConfig  *EmailHandlerConfig
	Bath           *databath.Bath
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

func (h *EmailHandler) Preview(requestTorch *torch.Request) {
	functionName := ""
	emailName := ""
	var id uint64

	err := requestTorch.UrlMatch(&functionName, &emailName, &id)
	if err != nil {
		log.Println(err)
		return
	}
	w := requestTorch.GetWriter()
	w.Header().Add("content-type", "text/html")
	emailConfig, ok := h.HandlerConfig.Templates[emailName]
	if !ok {
		log.Println("Template not found")
		return
	}
	err = h.TemplateWriter.Write(w, requestTorch, &emailConfig, id)
	if err != nil {
		log.Println(err)
		return
	}

}

func (h *EmailHandler) Send(requestTorch *torch.Request) {
	functionName := ""
	emailName := ""
	var id uint64
	recipient := ""
	notes := ""

	err := requestTorch.UrlMatch(&functionName, &emailName, &id, &recipient, &notes)
	if err != nil {
		requestTorch.DoError(err)
		return
	}
	w := bytes.Buffer{}
	emailConfig, ok := h.HandlerConfig.Templates[emailName]
	if !ok {
		requestTorch.DoErrorf("Template %s not found", emailName)
		return
	}
	err = h.TemplateWriter.Write(&w, requestTorch, &emailConfig, id)
	if err != nil {
		log.Println(err)
		return
	}
	str := w.String()
	parts := strings.SplitN(str, "\n", 2)
	if len(parts) != 2 {
		log.Println("PARTS LENGH != 2")
		requestTorch.DoErrorf("Email template didn't have a newline, could not extract a subject")
		return
	}

	notes = strings.Replace(notes, "\n", "<br/>", -1)
	html := parts[1]
	html = strings.Replace(html, "--- NOTES HERE ---", notes, 1)

	email := Email{
		Sender:    h.HandlerConfig.From,
		Recipient: recipient,
		Subject:   parts[0],
		Html:      html,
	}

	err = h.Sender.Send(&email)
	if err != nil {
		requestTorch.DoError(err)
		return
	}
	requestTorch.Write("Email Sent Successfully")

}
