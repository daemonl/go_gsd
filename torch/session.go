package torch

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
)

type Session struct {
	Key   *string // Points to the actual key
	User  *User
	Store *SessionStore
	Flash []FlashMessage
}

type SessionStore struct {
	sessions map[string]*Session
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
func InMemorySessionStore() *SessionStore {
	ss := SessionStore{
		sessions: make(map[string]*Session),
	}
	return &ss
}

func (ss *SessionStore) NewSession() *Session {
	randBytes := make([]byte, 128, 128)
	_, _ = rand.Reader.Read(randBytes)
	keyString := hex.EncodeToString(randBytes)
	session := Session{
		Key:   &keyString,
		Store: ss,
	}
	ss.sessions[keyString] = &session
	return &session
}

func (ss *SessionStore) GetSession(key string) (*Session, error) {
	sess, ok := ss.sessions[key]
	if !ok {
		fmt.Printf("Session Not Found: %s\n", key)
		return nil, errors.New("No session with that key")
	}
	return sess, nil
}

func (request *Request) End() {

}

func (request *Request) Write(content string) {
	request.writer.Write([]byte(content))
}

func (request *Request) Writef(format string, params ...interface{}) {
	request.Write(fmt.Sprintf(format, params...))
}

func (request *Request) PostValueString(name string) string {
	return request.raw.PostFormValue(name)
}

func (request *Request) Redirect(to string) {
	http.Redirect(request.writer, request.raw, to, 302)
}

func (request *Request) NewSession(store *SessionStore) {
	request.Session = store.NewSession()
	request.writer.Header().Add("Set-Cookie", fmt.Sprintf("gsd_session=%s", *request.Session.Key))
	//log.Printf("Generate New Session %s", *request.Session.Key)
}
