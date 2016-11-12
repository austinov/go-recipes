package view

const (
	ChatMessage = iota
	InfoMessage
	ErrorMessage
	TailMessage
)

var MessageKinds = map[byte]string{
	ChatMessage:  "",
	InfoMessage:  "*** Info ***",
	ErrorMessage: "*** Error ***",
	TailMessage:  "",
}

type View interface {
	ViewMessage(kind byte, from, message string)
	UpdatePeers(p []string)
	Show() <-chan struct{}
	Quit()
}
