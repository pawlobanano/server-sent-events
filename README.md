# server-sent-events
Proof of concept for Server-Sent Events (SSE) - a server push technology enabling a client to receive automatic updates from a server via an HTTP connection, and describes how servers can initiate data transmission towards clients once an initial client connection has been established.

## Example

**Run HTTP server**
```sh
go run main.go
```

**Subscribe**
```
curl -N http://localhost:8080/subscribe?topic=news
```

**Publish message on topic**
```
curl "http://localhost:8080/publish?topic=news&message=Hello%20world"
```
