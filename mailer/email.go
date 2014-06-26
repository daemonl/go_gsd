package mailer

type email struct {
	recipient string
	sender    string
	subject   string
	html      string
}

func (e *email) Recipient() string {
	return e.recipient
}
func (e *email) Sender() string {
	return e.sender
}
func (e *email) Subject() string {
	return e.subject
}
func (e *email) HTML() string {
	return e.html
}
