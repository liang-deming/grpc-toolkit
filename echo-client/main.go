// echo-client は Echo gRPC サービスを呼び出すサンプルクライアントのエントリポイントである。
//
// 接続先は直接の host:port ではなく、カスタムスキーム myscheme とサービス名 myecho を組み合わせた
// target（例: grpc.Dial("myscheme:///myecho", ...)）で、名前解決は Name サービスと GetNameResolver 経由で行う。
//
// 接続は client_pool によりプールされ、Unary 呼び出しを複数回行うデモになっている。
package main

import (
	"flag"
	"fmt"
	"grpcv2/echo"
	"grpcv2/echo-client/client"
	"grpcv2/echo-client/client_pool"
	"log"
	"time"

	"google.golang.org/grpc"
)

var (
	// addr は直接ダイアル用のフラグ（現在コメントアウトされた Dial パス向け）。既定は未使用に近い。
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

// getOptions は grpc.Dial に渡すクライアント設定を構成する。
//
//   - mTLS: クライアント証明書 + CA によるサーバー検証
//   - Unary / Stream インターセプタ: PerRPC 資格情報の有無をデバッグ出力するサンプル
//   - WithPerRPCCredentials: OAuth2 静的トークンを各 RPC に載せる（サーバー側 interceptor と一致させる）
//   - カスタム Resolver: 名前サービスから実アドレスを取得
//   - keepalive クライアントパラメータ
func getOptions() []grpc.DialOption {
	var opts []grpc.DialOption
	//opts = append(opts, client.GetTlsOpt())
	opts = append(opts, client.GetMTlsOpt())
	opts = append(opts, grpc.WithUnaryInterceptor(client.UnaryInterceptor))
	opts = append(opts, grpc.WithStreamInterceptor(client.StreamInterceptor))
	opts = append(opts, client.GetAuth(client.FetchToken()))
	opts = append(opts, client.GetNameResolver(client.NewNameServer("localhost:60051")))
	opts = append(opts, client.GetKeepaliveOpt())
	return opts
}

func main() {
	//flag.Parse()
	//conn, err := grpc.Dial(*addr, getOptions()...)
	// プロトコルとサービス名でネームサービス経由に解決し、サーバーへ接続する
	//conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", client.MyScheme, client.MyServiceName), getOptions()...)

	pool, err := client_pool.GetPool(fmt.Sprintf("%s:///%s", client.MyScheme, client.MyServiceName), getOptions()...)
	if err != nil {
		log.Fatal(err)
	}
	conn := pool.Get()
	defer pool.Put(conn)

	c := echo.NewEchoServiceClient(conn)
	client.CallUnary(c)
	time.Sleep(8 * time.Second)
	client.CallUnary(c)
	time.Sleep(8 * time.Second)
	client.CallUnary(c)
	time.Sleep(8 * time.Second)
	client.CallUnary(c)
	time.Sleep(8 * time.Second)
	client.CallUnary(c)
	time.Sleep(8 * time.Second)
	//client.CallServerStream(c)
	//client.CallClientStream(c)
	//client.CallBidirectional(c)

}
