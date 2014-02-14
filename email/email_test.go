package email

import (
	"fmt"

	"testing"
)

func TestSend(t *testing.T) {

	conf := SmtpConfig{
		ServerAddress: "smtp.gmail.com",
		EhloAddress:   "smtp.gmail.com",
		ServerPort:    "587",
		Username:      "",
		Password:      "",
	}
	s := Sender{
		Config: &conf,
	}
	email := Email{
		Sender:    "",
		Recipient: "",
		Subject:   "This Is a Subject",
		Html:      "Hello World",
	}

	err := s.Send(&email)
	fmt.Println("DO")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("END")
}
