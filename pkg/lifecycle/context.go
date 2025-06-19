package lifecycle

import (
	"context"
	"time"
)

type CancellableContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// DeriveContext creates a new cancellable Context derived from the provided parent context.
func DeriveContext(parent context.Context) *CancellableContext {
	ctx, cancel := context.WithCancel(parent)
	return &CancellableContext{ctx, cancel}
}

func (c *CancellableContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *CancellableContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *CancellableContext) Err() error {
	return c.ctx.Err()
}

func (c *CancellableContext) Value(key any) any {
	return c.ctx.Value(key)
}

func (c *CancellableContext) Cancel() {
	c.cancel()
}
