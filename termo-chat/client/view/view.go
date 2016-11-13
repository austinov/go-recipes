package view

const (
	ChatMessage = iota
	InfoMessage
	ErrorMessage
	TailMessage
)

var MessageKinds = map[byte]string{
	ChatMessage:  "",
	InfoMessage:  "[Info]",
	ErrorMessage: "[Error]",
	TailMessage:  "",
}

type View interface {
	ViewRoom(id string)
	ViewMessage(kind byte, from, message string)
	UpdatePeers(peers []string)
	Show() <-chan struct{}
	Quit()
}
