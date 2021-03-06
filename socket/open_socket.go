package socket

import (
	"database/sql"
	"encoding/json"
	"github.com/daemonl/go_gsd/shared"
	"golang.org/x/net/websocket"
	"log"
)

type OpenSocket struct {
	session shared.ISession
	ws      *websocket.Conn
	Sender  chan SocketMessage
	Closer  chan bool
	closed  bool
	Manager *Manager
	UID     uint
}

type socketError struct {
	Message string `json:"message"`
}

func (os *OpenSocket) DB() (*sql.DB, error) {
	return os.Manager.GetDatabase(os.session)
}

func (os *OpenSocket) Close() {

	if os.closed {
		log.Printf("S:%d RE CLOSE\n", os.UID)
		return
	}
	log.Printf("S:%d CLOSE\n", os.UID)
	os.closed = true
	os.Closer <- true
	os.ws.Close()
	var indexOfThisSocket *int
	for i, s := range os.Manager.OpenSockets {
		if s == os {
			indexOfThisSocket = &i
			break
		}
	}
	if indexOfThisSocket == nil {
		return
	}

	// Swap the last item and this item, then remove the last item
	os.Manager.OpenSockets[*indexOfThisSocket] = os.Manager.OpenSockets[len(os.Manager.OpenSockets)-1]
	os.Manager.OpenSockets = os.Manager.OpenSockets[0 : len(os.Manager.OpenSockets)-1]
}
func (os *OpenSocket) Wait() {
	for {
		select {
		case msg := <-os.Sender:
			os.ws.Write([]byte("BEGIN|" + msg.GetFunctionName() + "|" + msg.GetResponseId()))
			msg.PipeMessage(os.ws)
			os.ws.Write([]byte("END|" + msg.GetFunctionName() + "|" + msg.GetResponseId()))
		case _ = <-os.Closer:
			return
		}
	}
}
func (os *OpenSocket) SendObject(functionName string, responseId string, object interface{}) {
	bytes, err := json.Marshal(object)
	if err != nil {
		os.SendError(responseId, err)
		return
	}
	m := StringSocketMessage{FunctionName: functionName, ResponseId: responseId, Message: string(bytes)}
	os.Sender <- &m
}
func (os *OpenSocket) SendRaw(sm SocketMessage) {
	os.Sender <- sm
}
func (os *OpenSocket) SendError(responseId string, err error) {
	errObject := socketError{
		Message: err.Error(),
	}
	bytes, _ := json.Marshal(errObject)
	m := StringSocketMessage{FunctionName: "error", ResponseId: responseId, Message: string(bytes)}
	log.Printf("S:%d SEND ERROR %s\n", os.UID, responseId)
	os.Sender <- &m
}
func (os *OpenSocket) Session() shared.ISession {
	return os.session
}
func (os *OpenSocket) Broadcast(functionName string, object interface{}) {
	os.Manager.Broadcast(functionName, object)
}

func (os *OpenSocket) GetContext() shared.IContext {
	return os.Session().User().GetContext()
}
