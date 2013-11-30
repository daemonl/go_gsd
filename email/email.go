package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type Sender struct {
	Config *SmtpConfig
	Auth   *smtp.Auth
}

type SmtpConfig struct {
	ServerAddress string
	EhloAddress   string
	ServerPort    string
	Username      string
	Password      string
}

type Email struct {
	Recipient string
	Sender    string
	Subject   string
	Html      string
}

func (s *Sender) Send(email *Email) error {

	recipients := make([]string, 1, 1)
	recipients[0] = email.Recipient
	headers := map[string]string{
		"To":           email.Recipient,
		"From":         email.Sender,
		"Reply-To":     email.Sender,
		"Subject":      email.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain",
	}

	buf := bytes.NewBuffer(nil)
	for k, v := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	buf.WriteString(email.Html)

	log.Printf("Dial %s:%s", s.Config.ServerAddress, s.Config.ServerPort)

	c, err := smtp.Dial(s.Config.ServerAddress + ":" + s.Config.ServerPort)
	if err != nil {
		log.Println(err)
		return err

	}
	defer c.Quit()

	log.Println("Connected")

	log.Println("EHLO")

	if err = c.Hello(s.Config.EhloAddress); err != nil {
		log.Println(err)
		return err
	}
	log.Println("EHLO Done")

	log.Println("Start TLS")
	tlsConfig := tls.Config{}
	if err = c.StartTLS(&tlsConfig); err != nil {
		log.Println(err)
		return err
	}
	log.Println("TLS Done")

	auth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.ServerAddress)

	log.Println("Begin Auth")
	if err = c.Auth(auth); err != nil {
		log.Println(err)
		return err
	}
	log.Println("End Auth")

	/*
		boundary := "f46d043c813270fc6b04c2d223da"

		buf.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\n\n")
		buf.WriteString("--" + boundary + "\n")
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n\n")
		docText.WriteTo(buf)
		buf.WriteString("\n\n--" + boundary + "\n")
		buf.WriteString("Content-Type: text/html; charset=utf-8\n\n")
		docHtml.WriteTo(buf)
		buf.WriteString("--" + boundary + "--\n")
	*/

	log.Printf("Sender %s", email.Sender)
	if err = c.Mail(email.Sender); err != nil {
		c.Reset()
		log.Println(err)
		return err
	}

	log.Println("Recipient")
	if err = c.Rcpt(email.Recipient); err != nil {
		c.Reset()
		log.Println(err)
		return err
	}

	log.Println("Data")

	w, err := c.Data()
	if err != nil {
		log.Println(err)
		return err

	}
	buf.WriteTo(w)
	w.Close()
	c.Reset()
	log.Println("DONE")

	return nil
}
