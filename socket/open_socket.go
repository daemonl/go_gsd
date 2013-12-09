package socket

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/daemonl/go_gsd/torch"
	"log"
)

type OpenSocket struct {
	Session *torch.Session
	ws      *websocket.Conn
	Sender  chan SocketMessage
	Closer  chan bool
	Manager *Manager
}

type socketError struct {
	Message string `json:"message"`
}

func (os *OpenSocket) Close() {
	os.Closer <- true
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
			log.Println("Chan Socket Closed")
			return
		}
	}
}
func (os *OpenSocket) SendObject(functionName string, responseId string, object interface{}) {
	bytes, _ := json.Marshal(object)
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
	os.Sender <- &m
}
func (os *OpenSocket) SendObjectToAll(functionName string, object interface{}) {
	bytes, _ := json.Marshal(object)
	m := StringSocketMessage{FunctionName: functionName, ResponseId: "", Message: string(bytes)}
	log.Println("BROADCAST " + functionName)
	log.Println(string(bytes))
	for i, otherSocket := range os.Manager.OpenSockets {
		log.Printf("Send to %d of %d\n", i+1, len(os.Manager.OpenSockets))
		go otherSocket.SendRaw(&m)
	}
	log.Println("END BROADCAST")
}
