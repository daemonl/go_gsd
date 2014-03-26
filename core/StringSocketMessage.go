package core

import (
	"io"
)

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
