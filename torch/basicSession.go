package torch

import (
	"fmt"
	"time"
)

type basicSession struct {
	key         *string // Points to the actual key
	user        User
	store       SessionStore
	flash       []FlashMessage
	loginTarget *string
	lastRequest time.Time
}

type FlashMessage struct {
	Severity string
	Message  string
}

func (s *basicSession) Key() *string {
	return s.key
}

func (s *basicSession) UserID() *uint64 {
	if s.user == nil {
		return nil
	}
	id := s.user.ID()
	return &id
}

func (s *basicSession) AddFlash(severity, format string, parameters ...interface{}) {
	fm := FlashMessage{
		Severity: severity,
		Message:  fmt.Sprintf(format, parameters...),
	}
	s.flash = append(s.flash, fm)
}

func (s *basicSession) ResetFlash() {
	s.flash = make([]FlashMessage, 0, 0)
}

func (s *basicSession) DisplayFlash() []FlashMessage {
	fm := s.flash
	s.ResetFlash()
	return fm
}

func (s *basicSession) LastRequest() time.Time {
	return s.lastRequest
}

func (s *basicSession) User() User {
	return s.user
}

func (s *basicSession) SetUser(user User) {
	s.user = user
}

func (s *basicSession) Broadcast(name string, val interface{}) {

}

func (s *basicSession) UpdateLastRequest() {
	s.lastRequest = time.Now()
}

func (s *basicSession) SessionStore() SessionStore {
	return s.store
}
