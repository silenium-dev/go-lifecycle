package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	parentCtx  context.Context
	mainCtx    *CancellableContext
	loggingCtx *CancellableContext
	signalChan chan os.Signal
	mainDone   chan struct{}

	main          func(ctx context.Context, loggingCtx context.Context) error
	cleanup       func(ctx context.Context, loggingCtx context.Context)
	immediateExit func()
}

/*
NewApplication creates a new Application with specified main, cleanup, and exit handlers.

main is the primary application logic function to be executed.
cleanup (optional) is a function to handle cleanup tasks when the application stops.
immediateExit (optional) is invoked on receiving a second termination signal during shutdown.

There are two contexts available:
  - ctx: main context, all application actions should use this, canceled before cleanup
  - loggingCtx: lives slightly longer than the main context, canceled after cleanup,
    dedicated for logging background routines that need to run during cleanup
*/
func NewApplication(
	main func(ctx context.Context, loggingCtx context.Context) error,
	cleanup func(ctx context.Context, loggingCtx context.Context),
	immediateExit func(),
) *Application {
	app := &Application{
		parentCtx:     context.Background(),
		mainCtx:       DeriveContext(context.Background()),
		loggingCtx:    DeriveContext(context.Background()),
		signalChan:    make(chan os.Signal, 1),
		mainDone:      make(chan struct{}, 1),
		main:          main,
		cleanup:       cleanup,
		immediateExit: immediateExit,
	}
	signal.Notify(
		app.signalChan,
		os.Interrupt, os.Kill,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
	)
	go app.handleSignal()
	return app
}

// Context returns the application's main context.
// This context is used as the primary driver for the application's operations.
func (app *Application) Context() context.Context {
	return app.mainCtx
}

// LoggingContext returns the application's dedicated logging context.
// In contrast to the main context, it stays active during cleanup
// to allow background logs to be delivered
func (app *Application) LoggingContext() context.Context {
	return app.loggingCtx
}

// Run executes the application's main function within the configured contexts.
// Returns an error if the application has already been canceled or an eventual error from main
func (app *Application) Run() error {
	if app.mainCtx.Err() != nil {
		return fmt.Errorf("main context canceled: %w", app.mainCtx.Err())
	}
	if app.loggingCtx.Err() != nil {
		return fmt.Errorf("logging context canceled: %w", app.loggingCtx.Err())
	}
	if app.main == nil {
		return fmt.Errorf("no main function provided")
	}
	if err := app.main(app.mainCtx, app.loggingCtx); err != nil {
		return fmt.Errorf("error running application main: %w", err)
	}
	app.mainDone <- struct{}{}
	return nil
}

func (app *Application) handleSignal() {
	<-app.signalChan
	go app.handleImmediateExit()
	app.mainCtx.Cancel()
	if app.cleanup != nil {
		cleanupCtx := DeriveContext(app.parentCtx)
		defer cleanupCtx.Cancel()
		app.cleanup(cleanupCtx, app.loggingCtx)
	}
	<-app.mainDone
	app.loggingCtx.Cancel()
}

func (app *Application) handleImmediateExit() {
	<-app.signalChan
	if app.immediateExit != nil {
		app.immediateExit()
	}
	os.Exit(1)
}
