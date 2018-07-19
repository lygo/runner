package runner

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

type Closer func() error
type Runner func() error

type App struct {
	Runners     []Runner
	Slams       []Closer
	Started     chan struct{}
	Done        chan int
	shutdowning chan struct{}

	down            *sync.WaitGroup
	onceShutdown    *sync.Once
	onceSetExitCode *sync.Once

	closed   int32
	exitCode uint8
	errs     chan error
}

func New() *App {
	return &App{
		Runners:         make([]Runner, 0),
		Slams:           make([]Closer, 0),
		Started:         make(chan struct{}),
		Done:            make(chan int, 1),
		down:            &sync.WaitGroup{},
		onceShutdown:    &sync.Once{},
		onceSetExitCode: &sync.Once{},
		shutdowning:     make(chan struct{}),
		errs:            make(chan error, 1),
		closed:          -1,
	}
}

func (app *App) setExitCode(code uint8) {
	app.onceSetExitCode.Do(func() {
		app.exitCode = code
	})
}

func (app *App) Run() {
	var (
		runnersLength = len(app.Runners)
		up            = &sync.WaitGroup{}
	)

	if len(app.Slams) == 0 {
		log.Println(`WARN: your app doesn't have functions for close`)
	}

	app.errs = make(chan error, runnersLength)

	app.down.Add(runnersLength)
	up.Add(runnersLength)

	for _, r := range app.Runners {
		go func(run Runner) {
			up.Done()
			defer app.down.Done()
			defer func() {
				if e := recover(); e != nil {
					app.errs <- errors.New(fmt.Sprint(e))
				}
			}()

			if err := run(); err != nil {
				app.errs <- err
			}

		}(r)
	}
	up.Wait()

	if app.Started != nil {
		close(app.Started)
	}

	// run catcher of error for shutdown application
	go func() {
		select {
		case <-app.shutdowning:
			return
		case err, ok := <-app.errs:
			app.setExitCode(1)
			if ok {
				// check shutdowning. may be it will be
				select {
				case <-app.shutdowning:
					log.Println("ERR:", err)
					return
				default:
				}

				log.Println("ERR:", err)
				app.Shutdown()
			}
			return
		}
	}()
}

func safelyCallCloser(fn Closer) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	err = fn()

	return err
}

func (app *App) Shutdown() {

	if atomic.LoadInt32(&app.closed) != -1 {
		select {
		case app.Done <- int(app.closed):
		default:
			log.Println(`WARN: channel "Done" don't listen`)
		}

		log.Println(`WARN: app already closed`)
		return
	}

	app.onceShutdown.Do(app.shutdown)
}

func (app *App) shutdown() {
	close(app.shutdowning)

	var err error

	for i := len(app.Slams) - 1; i >= 0; i -= 1 {
		if err = safelyCallCloser(app.Slams[i]); err != nil {
			app.setExitCode(1)
			log.Println("ERR:", err)
		}
	}

	app.down.Wait()
	close(app.errs)

	for err = range app.errs {
		if err != nil {
			log.Println("ERR:", err)
		}
	}

	if err != nil || app.exitCode == 1 {
		atomic.AddInt32(&app.closed, 2)
		app.Done <- 1
	} else {
		atomic.AddInt32(&app.closed, 1)
		app.Done <- 0
	}
}
