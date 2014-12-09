package torch

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/daemonl/go_gsd/shared"
)

type basicSession struct {
	key         *string // Points to the actual key
	user        shared.IUser
	store       shared.ISessionStore
	flash       []shared.FlashMessage
	loginTarget *string
	lastRequest time.Time
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
	fm := shared.FlashMessage{
		Severity: severity,
		Message:  fmt.Sprintf(format, parameters...),
	}
	s.flash = append(s.flash, fm)
}

func (s *basicSession) ResetFlash() {
	s.flash = make([]shared.FlashMessage, 0, 0)
}

func (s *basicSession) DisplayFlash() []shared.FlashMessage {
	fm := s.flash
	s.ResetFlash()
	return fm
}
func (s *basicSession) Flash() []shared.FlashMessage {
	return s.DisplayFlash()
}

func (s *basicSession) LastRequest() time.Time {
	return s.lastRequest
}

func (s *basicSession) User() shared.IUser {
	return s.user
}

func (s *basicSession) SetUser(user shared.IUser) {
	s.user = user
}

func (s *basicSession) Broadcast(name string, val interface{}) {
	s.store.Broadcast(name, val)
}

func (s *basicSession) UpdateLastRequest() {
	s.lastRequest = time.Now()
}

func (s *basicSession) SessionStore() shared.ISessionStore {
	return s.store
}

func (s *basicSession) GetDatabaseConnection() (*sql.DB, error) {
	return s.store.GetDatabaseConnectionForSession(s)
}

func (s *basicSession) ReleaseDB(db *sql.DB) {
	s.store.ReleaseDatabaseConnection(db)
}
