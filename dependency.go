package pool

import (
	"time"

	"github.com/DoNewsCode/core/di"
)

// Providers provide a *pool.Pool to the core.
func Providers(options ...ProviderOptionFunc) di.Deps {
	return di.Deps{newPool(options...)}
}

// ProviderOptionFunc is the functional option to Providers.
type ProviderOptionFunc func(pool *Pool)

// WithConcurrency sets the maximum concurrency for the pool.
func WithConcurrency(concurrency int) ProviderOptionFunc {
	return func(pool *Pool) {
		pool.concurrency = concurrency
	}
}

// WithTimeout sets the timeout for the pool. The timeout is on per job basis. If
// the job's execution surpass its timeout, the async context will be cancelled.
func WithTimeout(timeout time.Duration) ProviderOptionFunc {
	return func(pool *Pool) {
		pool.timeout = timeout
	}
}

// WithShutdownEvents instructs the pool to exit at the given events. This
// matters because we want to make sure the pool shuts down gracefully. So we
// only shut down the pool after no more jobs are coming into the pool. By
// default, the pool shuts down when both OnHTTPServerShutdown and
// OnGRPCServerShutdown are triggered. That means it will not lose any jobs
// dispatched in http and gRPC handler. If the primary server is the gRPC server, this
// option can be set to:
//   WithShutdownEvents(core.OnGRPCServerShutdown)
// WithShutdownEvents accepts a variadic parameter. If more than one event is
// passed into the option, the pool waits until all of them are triggered.
//   WithShutdownEvents(core.OnHTTPServerShutdown, core.OnGRPCServerShutdown, OnMyCustomEvent)
func WithShutdownEvents(events ...interface{}) ProviderOptionFunc {
	return func(pool *Pool) {
		pool.shutdownEvents = events
	}
}
