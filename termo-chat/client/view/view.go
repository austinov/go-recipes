package view

type View interface {
	ReceiveMessage(from, message string)
	UpdatePeers(p []string)
	Show() <-chan struct{}
	Quit()
}
