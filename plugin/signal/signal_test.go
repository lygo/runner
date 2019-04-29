package signal

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/lygo/runner"
)

func makeNotifyer(sink <-chan os.Signal) notifyer {
	return func(c chan<- os.Signal, sigs ...os.Signal) {
		go func() {
			for sig := range sink {
				if sig == nil {
					return
				}

				println("got signal", sig.String())

				for _, expSig := range sigs {
					if expSig.String() == sig.String() {
						c <- sig
					}
				}
			}
		}()
	}
}

func makeTestApp() *runner.App {
	app := runner.New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	app.Runners = append(app.Runners, func() error {
		<-ctx.Done()
		return nil
	})

	app.Runners = append(app.Runners, func() error {
		<-ctx.Done()
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		cancel()
		return nil
	})
	return app
}

func TestWithStopOnSignal(t *testing.T) {
	app := makeTestApp()

	signaler := make(chan os.Signal, 0)
	defer close(signaler)

	notify = makeNotifyer(signaler)

	RegisterShutdownBySignals(app)

	app.Run()

	go func() {
		signaler <- syscall.SIGTERM
	}()

	code := <-app.Done
	if code != 0 {
		t.Errorf("expected code 0, got %d", code)
	}
}
