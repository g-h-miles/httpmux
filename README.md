# HttpMux

A high-performance HTTP request router that combines the speed of [httprouter](https://github.com/julienschmidt/httprouter) with full compatibility with Go's standard `net/http` patterns.

**Built on httprouter's proven foundation** - HttpMux preserves 90% of the original httprouter codebase, including its blazing-fast radix tree routing algorithm and precedence rules, while adding standard library compatibility.

## Why HttpMux?

HttpMux bridges the gap between performance and compatibility:

- **🚀 httprouter Performance**: Built on httprouter's proven radix tree implementation with minimal overhead
- **🔧 Standard Compatible**: Drop-in replacement for `http.ServeMux` patterns
- **📝 Standard Wildcards**: Uses Go 1.22+ `{name}` and `{name...}` syntax instead of `:name` and `*name`
- **🎯 Standard Handlers**: Works with `http.Handler` and `http.HandlerFunc` directly
- **📍 PathValue Support**: Full compatibility with `request.PathValue()` from Go 1.22+
- **🔄 Easy Migration**: Familiar API for developers coming from `http.ServeMux`
- **⚡ httprouter Routing**: Same routing logic and precedence rules as httprouter (no `{$}` needed)

## Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    "log"

    "github.com/g-h-miles/httpmux"
)

func main() {
    router := httpmux.NewServeMux() // or httpmux.New()

    // Standard http.HandlerFunc - no third parameter needed!
    router.HandleFunc("GET", "/users/{id}", func(w http.ResponseWriter, r *http.Request) {
        userID := r.PathValue("id")  // Standard Go 1.22+ PathValue
        fmt.Fprintf(w, "User ID: %s", userID)
    })

    // Standard http.Handler interface works too
    router.Handle("POST", "/users", http.HandlerFunc(createUser))

    // Convenient method shortcuts
    router.GET("/health", healthCheck)
    router.POST("/webhook", handleWebhook)

    log.Fatal(http.ListenAndServe(":8080", router))
}

func createUser(w http.ResponseWriter, r *http.Request) {
    // Your handler code here
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    // Your webhook handler
}
```

## Standard Wildcard Patterns

HttpMux uses Go 1.22+ standard wildcard syntax:

```go
// Named parameters
router.GET("/users/{id}", userHandler)           // matches /users/123
router.GET("/posts/{category}/{id}", postHandler) // matches /posts/tech/456

// Catch-all parameters
router.GET("/files/{filepath...}", fileHandler)   // matches /files/docs/readme.txt

// Access parameters using standard PathValue
func userHandler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    category := r.PathValue("category")
    filepath := r.PathValue("filepath")
}
```

## Migration from http.ServeMux

HttpMux is designed to be a drop-in replacement with method routing:

```go
// Before (http.ServeMux)
mux := http.NewServeMux()
mux.HandleFunc("/users/{id}", userHandler)

// After (HttpMux)
router := httpmux.NewServeMux()
router.HandleFunc("GET", "/users/{id}", userHandler)  // Just add the method!
```

## Migration from httprouter

Migrating from the original httprouter is straightforward:

```go
// Before (httprouter)
func userHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    id := ps.ByName("id")
}
router.GET("/users/:id", userHandler)

// After (HttpMux)
func userHandler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Standard Go method
}
router.GET("/users/{id}", userHandler)  // Standard wildcard syntax
```

## API Reference

### Router Methods

```go
// Create new router
router := httpmux.NewServeMux() // or httpmux.New()

// Register handlers
router.Handle(method, path, handler)           // http.Handler
router.HandleFunc(method, path, handlerFunc)   // http.HandlerFunc

// HTTP method shortcuts
router.GET(path, handlerFunc)
router.POST(path, handlerFunc)
router.PUT(path, handlerFunc)
router.PATCH(path, handlerFunc)
router.DELETE(path, handlerFunc)
router.HEAD(path, handlerFunc)
router.OPTIONS(path, handlerFunc)

// Static file serving
router.ServeFiles("/static/{filepath...}", http.Dir("./public"))

// Manual route lookup
handler, found := router.Lookup(method, path)
```

### Configuration Options

```go
router := httpmux.NewServeMux() // or httpmux.New()

// Automatic trailing slash redirects (default: true)
router.RedirectTrailingSlash = true

// Automatic path cleaning and case-insensitive redirects (default: true)
router.RedirectFixedPath = true

// Automatic METHOD not allowed responses (default: true)
router.HandleMethodNotAllowed = true

// Automatic OPTIONS responses (default: true)
router.HandleOPTIONS = true

// Custom handlers
router.NotFound = http.HandlerFunc(custom404)
router.MethodNotAllowed = http.HandlerFunc(custom405)
router.PanicHandler = customPanicHandler
```

## Performance

HttpMux maintains httprouter's exceptional performance with minimal overhead. Based on the [go-http-routing-benchmark](https://github.com/julienschmidt/go-http-routing-benchmark):

| Router      | Operations | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
| ----------- | ---------- | ------------ | ------------- | ----------------------- |
| HttpRouter  | 95,911     | 10,829       | 0             | 0                       |
| HttpMux     | 87,728     | 12,815       | 0             | 0                       |
| StandardMux | 13,772     | 85,506       | 20,336        | 842                     |

_Benchmarks run on Apple M2 Max CPU_

## Compatibility

- **Go Version**: Requires Go 1.22+ for `PathValue` support
- **Standard Library**: 100% compatible with `net/http` patterns
- **Middleware**: Works with any middleware expecting `http.Handler`
- **Testing**: Easy to test with `httptest` package

## Why Not Just Use http.ServeMux?

Go 1.22's `http.ServeMux` added wildcard support, but HttpMux offers:

1. **Better Performance**: Radix tree vs linear matching
2. **Method Routing**: Built-in HTTP method support
3. **More Features**: Trailing slash handling, case-insensitive matching
4. **Proven Stability**: Based on battle-tested httprouter foundation

## Contributing

Contributions are welcome! This project maintains compatibility with both Go's standard library patterns and httprouter's performance characteristics.

## License

BSD 3-Clause License (same as original httprouter)
