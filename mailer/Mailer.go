package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/daemonl/go_gsd/shared"
)

type Mailer struct {
	Config *SmtpConfig
}

type SmtpConfig struct {
	ServerAddress      string  `json:"serverAddress"`
	EhloAddress        string  `json:"ehloAddress"`
	ServerPort         string  `json:"port"`
	Username           string  `json:"username"`
	Password           string  `json:"password"`
	DevOverrideAddress *string `json:"devOverrideAddress"`
}

func (s *Mailer) SendMailSimple(to string, subject string, body string) {
	email := &shared.Email{
		Recipient: to,
		Subject:   subject,
		HTML:      body,
		Sender:    s.Config.Username,
	}
	s.SendMail(email)
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

func (s *Mailer) SendMailFromResponse(response shared.IResponse, recipientsRaw string, notes string) error {
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
			Sender:    s.Config.Username,
			Recipient: recipient,
			Subject:   subject,
			HTML:      html,
		}

		err := s.SendMail(email)
		if err != nil {
			return err
		}

	}
	return nil

}

func (s *Mailer) SendMail(email *shared.Email) error {

	if s.Config.DevOverrideAddress != nil && len(*s.Config.DevOverrideAddress) > 0 {
		email.HTML = email.Recipient + "<hr>" + email.HTML
		email.Recipient = *s.Config.DevOverrideAddress
	}

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
		tlsConfig := tls.Config{
			ServerName: s.Config.ServerAddress,
		}
		if err = c.StartTLS(&tlsConfig); err != nil {
			err = fmt.Errorf("SMTP TLS Error: %s", err.Error())
			log.Println(err)
			return err
		}

		auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.ServerAddress)

		if err = c.Auth(auth); err != nil {
			err = fmt.Errorf("SMTP Auth error: %s", err.Error())
			log.Println(err)
			return err
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

	writer, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMPT Data Error: %s", err.Error())
	}

	mw := multipart.NewWriter(writer)

	headers := map[string]string{
		"From":         email.Sender,
		"To":           email.Recipient,
		"Subject":      email.Subject,
		"MIME-Version": "1.0",
		"Content-Type": `multipart/mixed; boundary="` + mw.Boundary() + `"`,
	}

	for key, val := range headers {
		fmt.Fprintf(writer, "%s: %s\n", key, val)
	}
	fmt.Fprintln(writer, "")

	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Add("Content-Type", "text/html")
	htmlPart, err := mw.CreatePart(htmlHeader)
	if err != nil {
		return err
	}
	fmt.Fprintln(htmlPart, email.HTML)

	mw.Close()
	writer.Close()
	c.Reset()
	log.Println("DONE")

	return nil
}
