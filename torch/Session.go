package torch

import (
	"fmt"
	"time"
)

type Session struct {
	Key         *string // Points to the actual key
	User        *User
	Store       *SessionStore
	Flash       []FlashMessage
	LoginTarget *string
	LastRequest time.Time
}

type FlashMessage struct {
	Severity string
	Message  string
}

func (s *Session) AddFlash(severity, format string, parameters ...interface{}) {
	fm := FlashMessage{
		Severity: severity,
		Message:  fmt.Sprintf(format, parameters...),
	}
	s.Flash = append(s.Flash, fm)
}

func (s *Session) ResetFlash() {
	s.Flash = make([]FlashMessage, 0, 0)
}
