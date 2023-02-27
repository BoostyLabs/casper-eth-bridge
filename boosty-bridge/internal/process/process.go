package process

import (
	"os"
	"os/signal"
	"syscall"
)

// OnSigInt fires in SIGINT or SIGTERM event (usually CTRL+C).
func OnSigInt(onSigInt func()) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		onSigInt()
	}()
}
