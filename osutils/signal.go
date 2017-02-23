package osutils

import (
	"os"
	"os/signal"
)

// SignalHandle executes the handler when any signal is received.
func SignalHandle(handler func(), sig ...os.Signal) {
	interrupter := make(chan os.Signal, 1)
	signal.Notify(interrupter, sig...)
	go func() {
		defer close(interrupter)
		<-interrupter
		handler()
		signal.Stop(interrupter)
	}()
}
