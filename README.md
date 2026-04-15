# grpc-toolkit

A Go sample workspace demonstrating [gRPC](https://grpc.io/) patterns: an **Echo** service (unary and streaming RPCs), a small **in-memory name service** for service discovery, **mTLS**, **OAuth2-style per-RPC credentials**, **interceptors**, **HTTP/2 keepalive**, a **custom name resolver**, and a simple **client connection pool**.

**Languages:** [English](README.md) · [日本語](README.ja.md)

## Requirements

- Go **1.24** or newer (see `go.mod`)
- TLS assets under `x509/` (server/client certificates and CAs used by `echo-server` and `echo-client`)

## Repository layout

| Path | Role |
|------|------|
| `echo/` | `echo.proto` and generated `*.pb.go` — Echo RPC definitions |
| `name/` | `name.proto` and generated `*.pb.go` — name registry RPCs |
| `echo-server/` | Echo service implementation, health check, registration with the name server |
| `echo-client/` | Client with custom resolver (`myscheme:///myecho`), mTLS, pool, sample calls |
| `name-server/` | In-memory name registry (`Register`, `GetAddress`, …) |
| `x509/` | PEM certificates for mTLS (paths are hard-coded in TLS helpers) |

## Features (high level)

- **EchoService**: unary plus server/client/bidirectional streaming (streaming demos read/write JPEG paths under `echo-server/file/` and `echo-client/file/`).
- **Name service**: logical name → list of addresses; echo-server registers itself on startup; echo-client resolves via a custom `resolver.Builder` (`myscheme`).
- **Security**: mutual TLS between client and server; metadata `Authorization: Bearer <token>` checked on the server (demo token `some-secret-token`).
- **Keepalive**: server `MaxConnectionAge` / idle settings may close connections periodically; the client may see resolver `ResolveNow` after reconnects.

## Running the stack

Start processes **in this order** so the client does not get **“name resolver error: produced zero addresses”** (the registry must contain `myecho` before the client dials).

1. **Name server** (default `:60051`):

   ```bash
   go run ./name-server
   ```

2. **Echo server** (default `:50056`; registers `myecho` → `localhost:<port>` with the name server at `localhost:60051`):

   ```bash
   go run ./echo-server
   ```

3. **Echo client** (dials `myscheme:///myecho`, resolves via name server, calls `UnaryEcho` in a loop):

   ```bash
   go run ./echo-client
   ```

Override the echo listen port:

```bash
go run ./echo-server -port 50057
```

Ensure the name server still matches what `echo-client` uses for `NewNameServer("localhost:60051")` and that registration uses the same port.

## Regenerating protobuf code

If you change `.proto` files, regenerate with your usual `protoc` / `buf` setup. Generated files in this repo include `echo.pb.go`, `echo_grpc.pb.go`, `name.pb.go`, `name_grpc.pb.go` (do not edit by hand).

## Notes

- **JPEG files**: streaming examples expect image paths such as `echo-client/file/client.jpg` and `echo-server/file/server.jpg`. The repo `.gitignore` may ignore `*.jpg`; add suitable files locally if you enable streaming calls.
- **Resolver demo**: `ResolveNow` in the custom resolver may replace addresses with fixed hosts for demonstration; trim logging of `*grpc.ClientConn` in production.
- **Deprecated API**: `grpc.Dial` is used in places; newer code may prefer `grpc.NewClient` with the same concepts.

## License

No license file is included in this repository; add one if you distribute the project.
