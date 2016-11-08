package handler

type Sender interface {
	SendMessage(message string) error
}
