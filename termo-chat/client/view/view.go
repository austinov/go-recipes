package view

const (
	ChatMessage = iota
	InfoMessage
	ErrorMessage
)

var MessageKinds = map[byte]string{
	ChatMessage:  "",
	InfoMessage:  "*** Info ***",
	ErrorMessage: "*** Error ***",
}

type View interface {
	ReceiveMessage(kind byte, from, message string)
	UpdatePeers(p []string)
	Show() <-chan struct{}
	Quit()
}
