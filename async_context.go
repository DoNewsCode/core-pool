package pool

import (
	"context"
	"time"
)

type asyncContext struct {
	cancelCtx context.Context
	valueCtx  context.Context
}

func (a asyncContext) Deadline() (time.Time, bool)       { return a.cancelCtx.Deadline() }
func (a asyncContext) Done() <-chan struct{}             { return a.cancelCtx.Done() }
func (a asyncContext) Err() error                        { return a.cancelCtx.Err() }
func (a asyncContext) Value(key interface{}) interface{} { return a.valueCtx.Value(key) }
