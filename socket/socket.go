package socket

import (
	"bufio"
	"code.google.com/p/go.net/websocket"
	"database/sql"
	"encoding/json"
	"github.com/daemonl/go_gsd/actions"
	"github.com/daemonl/go_gsd/shared"
	"io"
	"log"
	"strings"
	"time"
)

var nextUID uint = 0

type Manager struct {
	handlers         map[string]Handler
	websocketHandler websocket.Handler
	sessionStore     shared.ISessionStore
	OpenSockets      []*OpenSocket
	GetDatabase      func(shared.ISession) (*sql.DB, error)
}

type SocketMessage interface {
	PipeMessage(io.Writer)
	GetFunctionName() string
	GetResponseId() string
}

type StringSocketMessage struct {
	Message      string
	FunctionName string
	ResponseId   string
}

func (ssm *StringSocketMessage) GetFunctionName() string {
	return ssm.FunctionName
}

func (ssm *StringSocketMessage) GetResponseId() string {
	return ssm.ResponseId
}

func (ssm *StringSocketMessage) PipeMessage(w io.Writer) {
	w.Write([]byte(ssm.Message))
}

type Handler interface {
	RequestDataPlaceholder() interface{}
	Handle(ac actions.Request, requestObject interface{}) (shared.IResponse, error)
}

func GetManager(sessionStore shared.ISessionStore) *Manager {
	m := Manager{
		handlers:     make(map[string]Handler),
		sessionStore: sessionStore,
		OpenSockets:  make([]*OpenSocket, 0, 0),
	}
	m.websocketHandler = websocket.Handler(m.listener)

	sessionStore.SetBroadcast(m.Broadcast)

	return &m
}

func (m *Manager) RegisterHandler(name string, handler Handler) {
	m.handlers[name] = handler
}

func (m *Manager) GetListener() *websocket.Handler {
	return &m.websocketHandler
}

func (manager *Manager) Broadcast(functionName string, object interface{}) {
	bytes, _ := json.Marshal(object)
	m := StringSocketMessage{FunctionName: functionName, ResponseId: "", Message: string(bytes)}
	log.Println("BROADCAST " + functionName)
	log.Println(string(bytes))
	for i, s := range manager.OpenSockets {
		log.Printf("Send to %d of %d\n", i+1, len(manager.OpenSockets))
		go s.SendRaw(&m)
	}
	log.Println("END BROADCAST")
}

// Echo the data received on the WebSocket.
func (m *Manager) listener(ws *websocket.Conn) {
	sessCookie, err := ws.Request().Cookie("gsd_session")
	if err != nil {
		log.Println(err)
		ws.Write([]byte("BEGIN|auth|"))
		ws.Write([]byte("{\"message\":\"Not Logged In\"}"))
		ws.Write([]byte("END|auth|"))
		time.Sleep(time.Second * 1)
		ws.Close()
		return
	}

	session, err := m.sessionStore.GetSession(sessCookie.Value)
	if err != nil {
		log.Println(err)
		ws.Close()
		return
	}

	if session.User == nil {
		log.Println("Socket opened for session with no user")
		ws.Close()
		return
	}

	os := OpenSocket{
		session: session,
		ws:      ws,
		Sender:  make(chan SocketMessage, 5),
		Closer:  make(chan bool),
		Manager: m,
		UID:     nextUID,
	}
	nextUID++
	m.OpenSockets = append(m.OpenSockets, &os)

	go os.Wait()
	defer os.Close()

	whoAmIString, _ := json.Marshal(session.User().WhoAmIObject())

	whoAmI := StringSocketMessage{
		Message:      string(whoAmIString),
		FunctionName: "whoami",
		ResponseId:   "",
	}
	os.Sender <- &whoAmI

	log.Printf("S:%d OPEN for user %d\n", os.UID, session.UserID())

	r := bufio.NewReader(ws)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Printf("S:%d Error reading line: %s\n", os.UID, err.Error())
			return
		}
		log.Printf("S:%d IN: %s", os.UID, line)
		os.session.UpdateLastRequest()
		m.parse(line, &os)
	}
}

func (m *Manager) parse(raw string, os *OpenSocket) {

	parts := strings.SplitN(raw, "|", 3)

	if len(parts) != 3 {
		log.Printf("Message should be functionName|requestObject|responseid.\n")
		return
	}
	functionName := parts[0]
	responseId := parts[1]

	handlerObj, ok := m.handlers[functionName]
	if !ok {
		log.Printf("No function named '%s'\n", functionName)
		return
	}

	requestObj := handlerObj.RequestDataPlaceholder()

	err := json.Unmarshal([]byte(parts[2]), requestObj)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := handlerObj.Handle(os, requestObj)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		os.SendError(responseId, err)
		return
	}

	os.SendObject("response", responseId, resp)

}
