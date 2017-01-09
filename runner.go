package runner

import (
	"google.golang.org/grpc/grpclog"
	"sync"
)

type Closer func() error
type Runner func() error

type App struct {
	Runners []Runner
	Slams   []Closer
	Started chan struct{}
}

func New() *App {
	return &App{
		Runners: make([]Runner, 0),
		Slams:   make([]Closer, 0),
		Started: make(chan struct{}, 0),
	}
}

func (app *App) Run() error {
	var (
		runnersLength = len(app.Runners)
		errC          = make(chan error, runnersLength)
		wg            = &sync.WaitGroup{}
	)

	wg.Add(runnersLength)

	for _, run := range app.Runners {
		go func() {
			wg.Done()
			// TODO: may be add defer for panic recover?
			errC <- run()
		}()
	}
	wg.Wait()
	if app.Started != nil {
		close(app.Started)
	}
	if err := <-errC; err != nil {
		app.Shutdown()
		close(errC)
		return err

	}
	return nil
}

func (app *App) Shutdown() {
	var err error

	for i := len(app.Slams) - 1; i >= 0; i -= 1 {
		if err = app.Slams[i](); err != nil {
			grpclog.Print(err)
		}
	}
}
