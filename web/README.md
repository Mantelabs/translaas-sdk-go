# web — framework integrations

Optional helpers for wiring `service.Service` into Go web applications. The root `web` package uses only the standard library; framework adapters are **separate modules** so the core SDK does not depend on Gin, Echo, or chi.

## Install

```bash
# stdlib net/http middleware
go get github.com/acuencadev/translaas-sdk-go

# optional framework modules
go get github.com/acuencadev/translaas-sdk-go/web/gin
go get github.com/acuencadev/translaas-sdk-go/web/echo
go get github.com/acuencadev/translaas-sdk-go/web/chi
```

## stdlib net/http

```go
apiClient, _ := client.New(client.Options{
    APIKey:  os.Getenv("TRANSLAAS_API_KEY"),
    BaseURL: os.Getenv("TRANSLAAS_BASE_URL"),
})
baseSvc, _ := service.New(apiClient, service.Options{Resolver: resolver})

mw, err := web.Middleware(web.DefaultMiddlewareOptions(baseSvc))
if err != nil {
    log.Fatal(err)
}

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    svc, ok := web.ServiceFromContext(r.Context())
    if !ok {
        http.Error(w, "missing translaas service", http.StatusInternalServerError)
        return
    }
    text, err := svc.T(r.Context(), "ui", "welcome")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    w.Write([]byte(text))
})

http.ListenAndServe(":8080", mw(mux))
```

## Gin

```go
mw, _ := translaasgin.Middleware(web.DefaultMiddlewareOptions(baseSvc))
r := gin.New()
r.Use(mw)
r.GET("/", func(c *gin.Context) {
    text, err := translaasgin.T(c, "ui", "welcome")
    // ...
})
```

## Echo

```go
mw, _ := translaasecho.Middleware(web.DefaultMiddlewareOptions(baseSvc))
e := echo.New()
e.Use(mw)
e.GET("/", func(c echo.Context) error {
    fn := translaasecho.TemplateFunc(c)
    text, err := fn("ui", "welcome")
    // ...
})
```

## chi

```go
mw, _ := translaaschi.Middleware(web.DefaultMiddlewareOptions(baseSvc))
r := chi.NewRouter()
r.Use(mw)
```

## Security

Translation strings are **not HTML-escaped** by the SDK. Use `html/template` for HTML output.

## Examples

Runnable samples live in the [translaas-all examples/go](https://github.com/acuencadev/translaas-all/tree/main/examples/go) folder.
