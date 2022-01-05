package pool

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/oklog/run"
)

func TestPool_Go(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	dispatcher := events.SyncDispatcher{}
	p := NewPool(WithConcurrency(1), WithShutdownEvents())(&dispatcher)
	go p.Go(context.Background(), func(asyncContext context.Context) {
		cancel()
	})
	p.Run(ctx)
}

func TestPool_Timeout(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	dispatcher := events.SyncDispatcher{}
	p := NewPool(WithTimeout(time.Second), WithConcurrency(1), WithShutdownEvents())(&dispatcher)
	go p.Go(context.Background(), func(asyncContext context.Context) {
		select {
		case <-asyncContext.Done():
			if asyncContext.Err() == nil {
				t.Fatalf("asyncContext should return err")
			}
			return
		case <-time.After(5 * time.Second):
			t.Fatal("should have timed out")
		}
	})
	p.Run(ctx)
}

func TestPool_FallbackToSyncMode(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	dispatcher := events.SyncDispatcher{}
	p := newPool(WithTimeout(time.Second), WithConcurrency(1), WithShutdownEvents())(&dispatcher)
	p.Run(ctx)

	var executed = make(chan struct{})
	go func() {
		// saturate the pool
		p.Go(ctx, func(asyncContext context.Context) {
			time.Sleep(time.Second)
		})
		// fallback to sync mode
		p.Go(ctx, func(asyncContext context.Context) {
			close(executed)
		})
	}()
	<-executed
}

func TestPool_contextValue(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	dispatcher := events.SyncDispatcher{}
	p := NewPool(WithConcurrency(1), WithShutdownEvents())(&dispatcher)
	key := struct{}{}
	requestContext := context.WithValue(context.Background(), key, "foo")
	go p.Go(requestContext, func(asyncContext context.Context) {
		if _, ok := asyncContext.Deadline(); !ok {
			t.Fatalf("asyncContext should have deadline set")
		}
		value := asyncContext.Value(key)
		if value != "foo" {
			t.Fatalf("want foo, got %s", value)
		}
	})
	p.Run(ctx)
}

func TestPool_ProvideRunGroup(t *testing.T) {
	t.Parallel()
	t.Run("run group should exit if no shutdown event is specified", func(t *testing.T) {
		dispatcher := events.SyncDispatcher{}
		p := NewPool(WithConcurrency(1), WithShutdownEvents())(&dispatcher)
		var group run.Group
		group.Add(func() error { return nil }, func(err error) {})
		p.ProvideRunGroup(&group)
		group.Run()
	})

	t.Run("run group should wait until all shutdown events", func(t *testing.T) {
		dispatcher := events.SyncDispatcher{}
		var fooEvent = "fooEvent"
		var barEvent = "barEvent"
		p := NewPool(WithConcurrency(1), WithShutdownEvents(fooEvent, barEvent))(&dispatcher)
		var group run.Group
		group.Add(func() error { return nil }, func(err error) {})
		p.ProvideRunGroup(&group)

		var final = make(chan struct{})

		go func() {
			group.Run()
			final <- struct{}{}
		}()

		select {
		case <-final:
			t.Fatal("group run should not exit now")
		case <-time.After(time.Second):
		}

		dispatcher.Dispatch(context.Background(), fooEvent, nil)

		select {
		case <-final:
			t.Fatal("group run should not exit now")
		case <-time.After(time.Second):
		}

		dispatcher.Dispatch(context.Background(), barEvent, nil)

		select {
		case <-final:
			return
		case <-time.After(4 * time.Second):
			t.Fatal("group should exit by now")
		}
	})
}
