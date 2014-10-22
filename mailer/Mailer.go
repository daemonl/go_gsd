package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/daemonl/go_gsd/shared"
)

type Mailer struct {
	Config *SmtpConfig
}

type SmtpConfig struct {
	ServerAddress string `json:"serverAddress"`
	EhloAddress   string `json:"ehloAddress"`
	ServerPort    string `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	DefaultSender string `json:"defaultSender"`
}

func (s *Mailer) SendSimple(to string, subject string, body string) {
	email := &shared.Email{
		Recipient: to,
		Subject:   subject,
		HTML:      body,
		Sender:    s.Config.DefaultSender,
	}
	s.Send(email)
}

func dropLine(in *string) (string, error) {
	log.Println("Drop Line")

	parts := strings.SplitN(*in, "\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("Could not extract line")
	}
	*in = parts[1]

	return parts[0], nil
}

func (s *Mailer) SendResponse(response shared.IResponse, recipientsRaw string, notes string) error {
	mailBuffer := &bytes.Buffer{}
	response.WriteTo(mailBuffer)
	html := mailBuffer.String()

	notes = strings.Replace(notes, "\n", "<br/>", -1)
	html = strings.Replace(html, "--- NOTES HERE ---", notes, 1)

	subject, err := dropLine(&html)
	if err != nil {
		return err
	}

	if recipientsRaw == "#inline" {
		recipientsRaw, err = dropLine(&html)
		if err != nil {
			return err
		}
	}

	recipients := strings.Split(recipientsRaw, ";")
	for _, recipient := range recipients {

		email := &shared.Email{
			Sender:    s.Config.DefaultSender,
			Recipient: recipient,
			Subject:   subject,
			HTML:      html,
		}

		err := s.Send(email)
		if err != nil {
			return err
		}

	}
	return nil

}

func (s *Mailer) Send(email *shared.Email) error {

	recipients := make([]string, 1, 1)
	recipients[0] = email.Recipient
	headers := map[string]string{
		"To":           email.Recipient,
		"From":         email.Sender,
		"Reply-To":     email.Sender,
		"Subject":      email.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html",
	}

	buf := bytes.NewBuffer(nil)
	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	buf.WriteString(email.HTML)

	log.Printf("Dial %s:%s", s.Config.ServerAddress, s.Config.ServerPort)

	c, err := smtp.Dial(s.Config.ServerAddress + ":" + s.Config.ServerPort)
	if err != nil {
		if s.Config.ServerPort == "9999" {
			log.Println("No dev email server is active")
			log.Printf(`
			MAILTO: %s
			SUBJECT: %s
			%s`, email.Recipient, email.Subject, email.HTML)
			return nil
		}
		log.Println(err)
		return err
	}
	defer c.Quit()

	// For Testing

	log.Println("Connected")

	log.Println("EHLO")

	if err = c.Hello(s.Config.EhloAddress); err != nil {
		log.Println(err)
		return err
	}
	log.Println("EHLO Done")

	if s.Config.ServerPort != "9999" {
		log.Println("Start TLS")
		tlsConfig := tls.Config{}
		if err = c.StartTLS(&tlsConfig); err != nil {
			return fmt.Errorf("SMTP TLS Error: %s", err.Error)
		}

		auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.ServerAddress)

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP Auth error: %s", err.Error())
		}

	}

	log.Printf("Sender %s", email.Sender)
	if err = c.Mail(email.Sender); err != nil {
		c.Reset()
		log.Println(err)
		return err
	}

	log.Printf("Recipient %s", email.Recipient)
	if err = c.Rcpt(email.Recipient); err != nil {
		c.Reset()
		log.Println(err)
		return err
	}

	log.Println("Data")

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMPT Data Error: %s", err.Error())
	}
	buf.WriteTo(w)
	w.Close()
	c.Reset()
	log.Println("DONE")

	return nil
}
