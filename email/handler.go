package email

import (
	"bytes"
	"errors"
	"github.com/daemonl/go_gsd/torch"
	"github.com/daemonl/go_gsd/view"
	"github.com/daemonl/databath"
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
	recipientRaw := ""
	notes := ""

	err := requestTorch.UrlMatch(&functionName, &emailName, &id, &recipientRaw, &notes)
	if err != nil {
		requestTorch.DoError(err)
		return
	}

	recipients := strings.Split(recipientRaw, ";")
	for _, recipient := range recipients {
		h.SendMailNow(emailName, id, strings.TrimSpace(recipient), notes, requestTorch)
	}

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
func (h *EmailHandler) SendMailNow(emailName string, id uint64, recipient string, notes string, requestTorch *torch.Request) {
	w := bytes.Buffer{}
	emailConfig, ok := h.HandlerConfig.Templates[emailName]
	if !ok {
		requestTorch.DoErrorf("Template %s not found", emailName)
		return
	}
	err := h.TemplateWriter.Write(&w, requestTorch, &emailConfig, id)
	if err != nil {
		log.Println(err)
		return
	}
	html := w.String()
	subject, err := dropLine(&html)
	if err != nil {
		if requestTorch != nil {
			requestTorch.DoError(err)
		} else {
			log.Println(err)
		}
		return
	}

	if recipient == "#inline" {
		recipient, err = dropLine(&html)
		if err != nil {
			if requestTorch != nil {
				requestTorch.DoError(err)
			} else {
				log.Println(err)
			}
			return
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
		if requestTorch != nil {
			requestTorch.DoError(err)
		} else {
			log.Println(err)
		}
		return
	}
	if requestTorch != nil {
		requestTorch.Write("Email Sent Successfully")
	} else {
		log.Println("Email Sent Successfully")
	}

}
