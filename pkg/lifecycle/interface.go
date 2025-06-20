package lifecycle

import "context"

type App interface {
	// Main is the primary application logic function to be executed.
	// ctx is the main context for all application logic
	// loggingCtx is a longer-lived context for background logging routines that is canceled after cleanup
	Main(ctx context.Context, loggingCtx context.Context) error
}

type CleanableApp interface {
	// Cleanup is a function to handle cleanup tasks when the application stops.
	// ctx is the main context for all application logic
	// loggingCtx is a longer-lived context for background logging routines that is canceled after cleanup
	Cleanup(ctx context.Context, loggingCtx context.Context)
}

type ImmediateExitApp interface {
	// ImmediateExit is invoked on receiving a second termination signal during shutdown.
	ImmediateExit()
}
