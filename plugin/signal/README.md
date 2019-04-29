# Signal

A simple way for triggered shutdown by system signals


# Usage

```go
package main

import (
	"github.com/lygo/runner"
	"github.com/lygo/runner/plugin/signal"
	"github.com/lygo/runner/examples/someapp"
)

func main() {
    app := someapp.New(someapp.NewConfig())
 
    // or it can be used inner someapp.New 
    signal.RegisterShutdownBySignals(app)
 
    app.Run()
    <-app.Started
    
    os.Exit(<-app.Done)	
}

```