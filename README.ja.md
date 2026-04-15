# grpc-toolkit

[gRPC](https://grpc.io/) の学習・検証用の Go ワークスペースです。**Echo** サービス（Unary とストリーミング）、**インメモリの名前サービス**によるサービス発見、**mTLS**、**OAuth2 風の Per-RPC 資格情報**、**インターセプタ**、**HTTP/2 keepalive**、**カスタム name resolver**、簡易な **クライアント接続プール** を含みます。

**言語:** [English](README.md) · [日本語](README.ja.md)

## 前提条件

- Go **1.24** 以降（`go.mod` 参照）
- `x509/` 以下の TLS 用 PEM（`echo-server` / `echo-client` の TLS ヘルパーが参照するパスに配置）

## ディレクトリ構成

| パス | 役割 |
|------|------|
| `echo/` | `echo.proto` と生成された `*.pb.go` — Echo の RPC 定義 |
| `name/` | `name.proto` と生成された `*.pb.go` — 名前レジストリの RPC |
| `echo-server/` | Echo の実装、ヘルスチェック、名前サービスへの自己登録 |
| `echo-client/` | カスタム resolver（`myscheme:///myecho`）、mTLS、プール、サンプル呼び出し |
| `name-server/` | インメモリの名前レジストリ（`Register`、`GetAddress` など） |
| `x509/` | mTLS 用の証明書（パスは TLS ヘルパーに固定） |

## 機能の概要

- **EchoService**: Unary に加え、サーバー／クライアント／双方向ストリーミング（ストリーム系は `echo-server/file/` と `echo-client/file/` 配下の JPEG パスを使用）。
- **名前サービス**: 論理サービス名 → アドレス一覧。echo-server 起動時に登録し、echo-client はカスタム `resolver.Builder`（`myscheme`）で解決する。
- **セキュリティ**: クライアントとサーバーの mTLS、メタデータ `Authorization: Bearer <token>` をサーバーで検証（デモ用トークン `some-secret-token`）。
- **Keepalive**: サーバーの `MaxConnectionAge` などにより接続が切れると、クライアント側で resolver の `ResolveNow` が走ることがある。

## 起動手順

クライアントが **「name resolver error: produced zero addresses」** にならないよう、**次の順**で起動してください（レジストリに `myecho` が載ってからクライアントが Dial する必要があるため）。

1. **名前サーバー**（既定 `:60051`）

   ```bash
   go run ./name-server
   ```

2. **Echo サーバー**（既定 `:50056`。名前サーバー `localhost:60051` に `myecho` → `localhost:<port>` を登録）

   ```bash
   go run ./echo-server
   ```

3. **Echo クライアント**（`myscheme:///myecho` で Dial し、名前サービスで解決して `UnaryEcho` を繰り返し呼び出し）

   ```bash
   go run ./echo-client
   ```

Echo の待受ポートを変える場合:

```bash
go run ./echo-server -port 50057
```

`echo-client` 内の `NewNameServer("localhost:60051")` と、サーバー側の登録先ポートと整合させてください。

## Protocol Buffers の再生成

`.proto` を変更した場合は、通常どおり `protoc` / `buf` で生成し直してください。本リポジトリでは `echo.pb.go`、`echo_grpc.pb.go`、`name.pb.go`、`name_grpc.pb.go` などが生成物のため、手編集は避けてください。

## 注意事項

- **JPEG**: ストリーミングのサンプルは `echo-client/file/client.jpg` や `echo-server/file/server.jpg` などを参照します。`.gitignore` で `*.jpg` を無視している場合は、ローカルでファイルを置いてください。
- **Resolver のデモ**: カスタム resolver の `ResolveNow` はデモ用にアドレスを固定値に差し替える処理が含まれる場合があります。本番では `*grpc.ClientConn` のそのままログ出力なども避けてください。
- **API**: 一部で `grpc.Dial` を使用しています。新規コードでは `grpc.NewClient` への移行を検討できます（概念は同じです）。

## ライセンス

リポジトリにライセンスファイルは含まれていません。配布する場合は適宜追加してください。
