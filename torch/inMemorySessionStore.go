package torch

import (
	"bufio"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type inMemorySessionStore struct {
	sessions          map[string]Session
	dumpFile          *string
	broadcast         func(string, interface{})
	getDatabaseMethod func(Session) (*sql.DB, error)
}

func InMemorySessionStore(dumpFile *string, loadUserById func(uint64) (User, error), getDatabaseMethod func(Session) (*sql.DB, error)) SessionStore {
	ss := inMemorySessionStore{
		sessions:          make(map[string]Session),
		dumpFile:          dumpFile,
		getDatabaseMethod: getDatabaseMethod,
	}

	go ss.StartExpiry()

	// Load any sessions which are stored in the session dump file.
	// This allows for sessions to persist when a normal restart happens.

	if dumpFile != nil {
		log.Println("=== HIDRATE SESSIONS ===")

		sessFile, err := os.Open(*dumpFile)
		if err != nil {
			log.Printf("Could not load sessions: %s\n", err.Error())
		} else {
			ss.loadSessions(sessFile, loadUserById)
			sessFile.Close()
		}
		log.Println("=== END SESSIONS ===")
	}

	return &ss
}

func (ss *inMemorySessionStore) DumpSessions() {
	if ss.dumpFile == nil {
		return
	}
	w, err := os.Create(*ss.dumpFile)
	if err != nil {
		log.Printf("Error dumping sessions: %s\n", err.Error())
	}
	//dumpFile
	for _, session := range ss.sessions {
		sKey := session.Key()
		sUser := session.UserID()
		if sKey != nil && sUser != nil {
			w.Write([]byte(fmt.Sprintf("%s|%d\n", *sKey, *sUser)))
		}
	}
}

func (ss *inMemorySessionStore) loadSessions(r io.Reader, loadUser func(uint64) (User, error)) {
	lr := bufio.NewReader(r)
	for {
		line, err := lr.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		userId, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			log.Println("Error loading session: %s\n", err)
			continue
		}
		s := &basicSession{
			key:         &parts[0],
			store:       ss,
			lastRequest: time.Now(),
		}
		user, err := loadUser(userId)
		if err != nil {
			log.Println("Error loading session: %s\n", err)
			continue
		}
		s.user = user
		log.Printf("Hidrated session for user %d\n", user.ID())
		ss.sessions[*s.Key()] = s

	}
}

func (ss *inMemorySessionStore) SetBroadcast(broadcast func(string, interface{})) {
	ss.broadcast = broadcast
}

func (ss *inMemorySessionStore) GetDatabaseConnectionForSession(session Session) (*sql.DB, error) {
	return ss.getDatabaseMethod(session)
}

func (ss *inMemorySessionStore) Broadcast(name string, data interface{}) {
	ss.broadcast(name, data)
}

func (ss *inMemorySessionStore) StartExpiry() {
	for {
		time.Sleep(time.Second * 10)

		for key, s := range ss.sessions {
			if time.Since(s.LastRequest()).Minutes() > 30 {
				log.Printf("Expire Session %s", key)
				delete(ss.sessions, key)
			}
		}
	}
}

func (ss *inMemorySessionStore) NewSession() (Session, error) {
	randBytes := make([]byte, 128, 128)
	_, _ = rand.Reader.Read(randBytes)
	keyString := hex.EncodeToString(randBytes)
	session := basicSession{
		key:         &keyString,
		store:       ss,
		lastRequest: time.Now(),
	}
	ss.sessions[keyString] = &session
	return &session, nil
}

func (ss *inMemorySessionStore) GetSession(key string) (Session, error) {
	sess, ok := ss.sessions[key]
	if !ok {
		fmt.Printf("Session Not Found: %s\n", key)
		return nil, nil //errors.New("No session with that key")
	}
	sess.UpdateLastRequest()
	return sess, nil
}
