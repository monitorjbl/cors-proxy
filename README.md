# CORS proxy

This is a really simple application designed to do one thing: serve as a proxy for requests requiring CORS.

```
Usage of cors-proxy:
  -address string
        The address the server will bind to (default "0.0.0.0")
  -auth string
        Enable Basic Auth on every request in user:pass form
  -debug
        Enable debug output
  -port string
        The port the server will bind to (default "8080")
```

# Building

```
go get 
```
