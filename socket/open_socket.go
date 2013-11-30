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

func (os *OpenSocket) Close() {
	os.Closer <- true
}
func (os *OpenSocket) Wait() {
	for {
		select {
		case msg := <-os.Sender:
			os.ws.Write([]byte("BEGIN|" + msg.GetFunctionName() + "|" + msg.GetResponseId()))
			msg.PipeMessage(os.ws)
			os.ws.Write([]byte("END|" + msg.GetFunctionName() + "|" + msg.GetResponseId()))
		case _ = <-os.Closer:
			log.Println("Close Chan Socket Wait")
			return
		}
	}
}
func (os *OpenSocket) SendObject(functionName string, responseId string, object interface{}) {
	bytes, _ := json.Marshal(object)
	m := StringSocketMessage{FunctionName: functionName, ResponseId: responseId, Message: string(bytes)}
	os.Sender <- &m
}

func (os *OpenSocket) SendObjectToAll(functionName string, object interface{}) {
	bytes, _ := json.Marshal(object)
	m := StringSocketMessage{FunctionName: functionName, ResponseId: "", Message: string(bytes)}
	log.Println("BROADCAST " + functionName)
	log.Println(string(bytes))
	for i, otherSocket := range os.Manager.OpenSockets {
		log.Printf("Send to %d\n", i)
		otherSocket.Sender <- &m
	}
	log.Println("END BROADCAST")
}
