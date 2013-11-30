package socket

import (
	"bufio"
	"log"

	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/daemonl/go_gsd/torch"
	"io"
	"strings"
)

type Manager struct {
	handlers         map[string]Handler
	websocketHandler websocket.Handler
	sessionStore     *torch.SessionStore
	OpenSockets      []*OpenSocket
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
	GetRequestObject() interface{}
	HandleRequest(os *OpenSocket, requestObject interface{}, responseId string)
}

func GetManager(sessionStore *torch.SessionStore) *Manager {
	m := Manager{
		handlers:     make(map[string]Handler),
		sessionStore: sessionStore,
		OpenSockets:  make([]*OpenSocket, 0, 0),
	}
	m.websocketHandler = websocket.Handler(m.listener)
	return &m
}

func (m *Manager) RegisterHandler(name string, handler Handler) {
	m.handlers[name] = handler
}

func (m *Manager) GetListener() *websocket.Handler {
	return &m.websocketHandler
}

// Echo the data received on the WebSocket.
func (m *Manager) listener(ws *websocket.Conn) {
	sessCookie, err := ws.Request().Cookie("gsd_session")
	if err != nil {
		fmt.Println(err)
		ws.Close()
		return
	}

	session, err := m.sessionStore.GetSession(sessCookie.Value)
	if err != nil {
		fmt.Println(err)
		ws.Close()
		return
	}

	os := OpenSocket{
		Session: session,
		ws:      ws,
		Sender:  make(chan SocketMessage, 5),
		Closer:  make(chan bool),
		Manager: m,
	}
	m.OpenSockets = append(m.OpenSockets, &os)

	go os.Wait()
	defer os.Close()

	whoAmIString, _ := json.Marshal(session.User)

	whoAmI := StringSocketMessage{
		Message:      string(whoAmIString),
		FunctionName: "whoami",
		ResponseId:   "",
	}
	os.Sender <- &whoAmI

	r := bufio.NewReader(ws)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("LINE IN: %s\n", line)
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

	handlerObj, ok := m.handlers[functionName]
	if !ok {
		log.Printf("No function named '%s'\n", functionName)
		return
	}

	requestObj := handlerObj.GetRequestObject()

	err := json.Unmarshal([]byte(parts[2]), requestObj)
	if err != nil {
		fmt.Println(err)
		return
	}

	handlerObj.HandleRequest(os, requestObj, parts[1])
}
