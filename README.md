The package pool is a generic solution for async job dispatching from web server. While Go natively
supports async jobs by using the keyword "go", but this may lead to several unwanted consequences.
Suppose we have a typical http handler:

```go
func Handle(req *http.Request, resp http.ResponseWriter) {}
```

If we dispatch async jobs using "go" like this:

```go
func Handle(req *http.Request, resp http.ResponseWriter) {
  go AsyncWork()
  resp.Write([]byte("ok"))
}
```

Let's go though all the disadvantages. First of all, the backpressure is lost. There is no way to
limit the maximum goroutine the handler can create. clients can easily flood the server. Secondly,
the graceful shutdown process is ruined. The http server can shutdown itself without losing any
request, but the async jobs created with "go" are not protected by the server. You will lose all
unfinished jobs once the server shuts down and program exits. lastly, the async job may want to
access the original request context, maybe for tracing purposes. The request context terminates at
the end of the request, so if you are not careful, the async jobs may be relying on a dead context.

Package pool creates a goroutine worker pool at beginning of the program, limits the maximum
concurrency for you, shuts it down at the end of the request without losing any async jobs, and
manages the context conversion for you.

Add the dependency to core:

```go
var c *core.C = core.New()
c.Provide(pool.Providers())
```

Then you can inject the pool into your http handler:

```go
type Handler struct {
  pool *pool.Pool
}

func (h *Handler) ServeHTTP(req *http.Request, resp http.ResponseWriter) {
  pool.Go(request.Context(), AsyncWork(asyncContext))
  resp.Write([]byte("ok"))
}
```


