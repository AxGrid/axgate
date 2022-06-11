AxGate Client
=============

Create client for your axgate-server

HTTP Tunnel
```go

//axgate.NewHTTPClient("<service>", "<axgate-server-tcp-address>", "<dst-address>")
err := axgate.NewHTTPClient("myservice", "localhost:9090", "http://ya.ru/")
```


HTTP Handler
```go

//axgate.NewHTTPHandlerClient("<service>", "<axgate-server-tcp-address>", <handler method>)
err := axgate.NewHTTPHandlerClient("myservice", "localhost:9090", handler)
```

AxGate Server
=============


```shell
axgate-server --tcp=":9090" --http=":80" --hosts="mydomain.com"
```