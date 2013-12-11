package torch

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
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

	go ss.StartExpiry()
	return &ss
}

func (ss *SessionStore) StartExpiry() {

	for {
		time.Sleep(time.Second * 10)

		log.Println("CHECK SESSION EXPIRY")
		for key, s := range ss.sessions {
			if time.Since(s.LastRequest).Minutes() > 1 {
				log.Printf("Expire Session %s", key)
				delete(ss.sessions, key)
			}
		}
	}

}

func (ss *SessionStore) NewSession() *Session {
	randBytes := make([]byte, 128, 128)
	_, _ = rand.Reader.Read(randBytes)
	keyString := hex.EncodeToString(randBytes)
	session := Session{
		Key:         &keyString,
		Store:       ss,
		LastRequest: time.Now(),
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
	sess.LastRequest = time.Now()
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
	c := http.Cookie{Name: "gsd_session", Path: "/", MaxAge: 3600, Value: *request.Session.Key}
	request.writer.Header().Add("Set-Cookie", c.String())
	//request.writer.Header().Add("Set-Cookie", fmt.Sprintf("gsd_session=%s", *request.Session.Key))
	//log.Printf("Generate New Session %s", *request.Session.Key)
}
