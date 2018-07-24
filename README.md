# Runner

The 'runner' can help you run your application with any a components and use one point for running and stopping them 

##	Approach

Basic usage will be looking something like that 

```go

package someapp

import (
	"net"
	
	"go.uber.org/zap"
	"github.com/lygo/runner"
)

type Config struct {
	HttpAddr string
	GrpcAddr string
	DBURI string
	OtherGrpcServiceEndpoint string
	GracefulTimeout time.Duration
}

// static validate
func (c Config) Validate() error {
	// just correct addresses and some database uri
	// or any params like timeout and retry config size
	return nil
}

func New(cfg Config)  (app *runner.App, err error) {
	// here our config are correct and we will do dynamic validation in the current environment 
    app = runner.New()
    
    defer func() {
        if err != nil {
            app.Shutdown()
        }
    }()
    
    var (
        grpcListener net.Listener
        apiListener  net.Listener
    )
    
    grpcListener, err = net.Listen(`tcp`, cfg.GrpcAddr)
    if err != nil {
        return
    }
    
    apiListener, err = net.Listen(`tcp`, cfg.HttpAddr)
    if err != nil {
        return
    }
    
    // or from config level and format
    var zapLogger *zap.Logger
    
    zapLogger, err = zap.NewDevelopment()
    if err != nil {
        return
    }
    zapLogger = zapLogger.Named(`service-name`)
    app.Slams = append(app.Slams, zapLogger.Sync)
    
    // also any metrics and tracing agent your can adding here 
    // with runnes and slams 
    
    // of course that code can move to contructor of your grpc service
    grpcSrv := grpc.NewServer(
        grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
            grpc_ctxtags.StreamServerInterceptor(),
            grpc_zap.StreamServerInterceptor(zapLogger),
            grpc_recovery.StreamServerInterceptor(),
        )),
        grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
            grpc_ctxtags.UnaryServerInterceptor(),
            grpc_zap.UnaryServerInterceptor(zapLogger),
            grpc_recovery.UnaryServerInterceptor(),
        )),
    )
    
    someServiceApb.RegisterServiceAServer(grpcSrv, serviceAImpl.New(cfg))
    someServiceBpb.RegisterServiceBServer(grpcSrv, serviceBImpl.New(cfg))
    
    app.Runners = append(app.Runners, func() error {
        err := grpcSrv.Serve(grpcListener)
        if err == grpc.ErrServerStopped {
            return nil
        }
        return err
    })
    
    app.Slams = append(app.Slams, func() error {
        grpcSrv.GracefulStop()
        return nil
    })
    
    // registry HTTP handlers
    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}
    err = someServiceBpb.RegisterServiceAHandlerFromEndpoint(context.Background(), mux, cfg.GrpcAddr, opts)
    if err != nil {
        return
    }

    err = someServiceApb.RegisterServiceBHandlerFromEndpoint(context.Background(), mux, cfg.GrpcAddr, opts)
    if err != nil {
        return
    }

    httpSrv := &http.Server{Handler: mux}
    
    app.Runners = append(app.Runners, func() error {
        err := httpSrv.Serve(apiListener)
        if err == http.ErrServerClosed {
            return nil
        }
        return err
    })

    app.Slams = append(app.Slams, func() error {
    	// also from time out 
        ctxDown, cancelGraceful := context.WithTimeout(context.Background(),cfg.GracefulTimeout)
        defer cancelGraceful()
        return httpSrv.Shutdown(ctxDown)
    })
    
    
    return
}

```

and your `main.go` file will be

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"someapp"
)

func main() {
	// fill and validate your config on `NewConfig` function
	cfg, err := someapp.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

    // get your app  
	app, err := someapp.New(*cfg)
	if err != nil {
		log.Fatal(err)
	}

    // run all functions from runners
	app.Run()

	ch := make(chan os.Signal, 10)
	signal.Notify(ch,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	defer close(ch)

    // wait when all will started
	<-app.Started

	// listen system channel for closing app
	go func() {
		sig, ok := <-ch
		if ok {
			log.Printf("shutdown process on %s system signal\n", sig)
		}
		app.Shutdown()
	}()

    // exist with correct status
    // if all slams with out error code code 0, in otherwise will 1
	os.Exit(<-app.Done)
}
```




 
