package runner

import (
	"errors"
	"testing"
)

func TestEmptyShutdown(t *testing.T) {
	// we can create new app
	// it is very easy %))
	app := New()
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()
	// and we can shutdown up without run
	app.Shutdown()
	// and retry
	app.Shutdown()
	// and try again
	app.Shutdown()
	// and again!
	app.Shutdown()

	if done := <-app.Done; done != 0 {
		t.Errorf("expected 0, got - %d", done)
	}
}

func TestErrorFirstsRunWithShutdown(t *testing.T) {
	app := New()
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	stopRunner := make(chan *struct{}, 0)
	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		return errors.New(`BOOM`)
	})

	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		return nil
	})
	app.Run()
	<-app.Started

	close(stopRunner)
	app.Shutdown()

	if done := <-app.Done; done != 1 {
		t.Errorf("expected 1, got - %d", done)
	}
}

func TestPanicErrorOnRun(t *testing.T) {
	app := New()
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	stopRunner := make(chan *struct{}, 0)
	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		panic(`BOOM`)
		return nil
	})

	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		return nil
	})
	app.Run()
	<-app.Started

	close(stopRunner)
	// wait of panic

	if done := <-app.Done; done != 1 {
		t.Errorf("expected 1, got - %d", done)
	}
}

func TestPanicErrorOnClose(t *testing.T) {
	app := New()
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	stopRunner := make(chan *struct{}, 0)
	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		return nil
	})

	app.Runners = append(app.Runners, func() error {
		<-stopRunner
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		close(stopRunner)
		return nil
	})

	app.Slams = append(app.Slams, func() error {
		panic(`BOOM`)
		return nil
	})
	app.Run()
	<-app.Started

	app.Shutdown()

	if done := <-app.Done; done != 1 {
		t.Errorf("expected 1, got - %d", done)
	}

	app.Shutdown()
}

func TestSpeedUpOnHighWay(t *testing.T) {
	// bad way of usage that pattern
	app := New()
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	app.Runners = append(app.Runners, func() error {
		return nil
	})

	app.Runners = append(app.Runners, func() error {
		return nil
	})

	app.Run()
	<-app.Started

	app.Shutdown()

	if done := <-app.Done; done != 0 {
		t.Errorf("expected 0, got - %d", done)
	}
}
