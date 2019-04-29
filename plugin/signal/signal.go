package signal

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lygo/runner"
)

var DefaultSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGKILL,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

type notifyer func(c chan<- os.Signal, sig ...os.Signal)

var notify notifyer = signal.Notify

// RegisterShutdownBySignals - start listen system signal on app.Run() and call app.Shutdown() on got termination system signals
func RegisterShutdownBySignals(app *runner.App, signals ...os.Signal) {
	if len(signals) == 0 {
		signals = DefaultSignals
	}

	ch := make(chan os.Signal, 10)
	notify(ch, signals...)
	done := make(chan struct{}, 0)

	app.Runners = append(app.Runners, func() error {
		defer close(done)
		defer close(ch)

		// listen system channel for closing app
		sig, ok := <-ch
		if ok {
			log.Printf("shutdown process on %s system signal\n", sig)
		}

		return nil
	})

	go func() {
		<-done
		app.Shutdown()
	}()
}
