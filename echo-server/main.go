// echo-server は gRPC の Echo サンプルサーバーのエントリポイントである。
//
// 主な役割:
//   - EchoService の実装を登録し、指定ポートで待ち受ける
//   - mTLS・Unary/Stream インターセプタ・keepalive などのサーバー側オプションを束ねる
//   - gRPC ヘルスチェックサービスを有効にし、ロードバランサ等からのプローブに応答できるようにする
//   - 別プロセスの Name サービス（既定 localhost:60051）へ、自サービス名と listen アドレスを登録する
//
// 終了処理: OS シグナル（Interrupt / Kill）を待ち、プロセス終了時にコンテキストがキャンセルされる。
package main

import (
	"context"
	"flag"
	"fmt"
	"grpcv2/echo"
	"grpcv2/echo-server/server"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	// port は Echo サーバーがリッスンする TCP ポート。フラグ -port で上書き可能。
	port = flag.Int("port", 50056, "The server port")
)

// getOptions は grpc.NewServer に渡すサーバー設定を組み立てる。
//
// 現在の構成:
//   - mTLS（クライアント証明書必須）: GetMTlsOpt
//   - Unary / Stream 用サーバーインターセプタ: OAuth2 風の Authorization ヘッダ検証
//   - keepalive の強制ポリシーとサーバー側パラメータ
//
// 単純な TLS のみにしたい場合は GetTlsOpt を使うコードがコメントとして残っている。
func getOptions() []grpc.ServerOption {
	var opts []grpc.ServerOption
	//opts = append(opts, server.GetTlsOpt())
	opts = append(opts, server.GetMTlsOpt())
	opts = append(opts, grpc.UnaryInterceptor(server.UnaryInterceptor))
	opts = append(opts, grpc.StreamInterceptor(server.StreamInterceptor))
	opts = append(opts, server.GetKeepaliveOpt()...)
	return opts
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
		return
	}

	s := grpc.NewServer(getOptions()...)
	echo.RegisterEchoServiceServer(s, &server.EchoServer{})

	// 標準の gRPC Health Checking Protocol。
	// 空のサービス名 "" に対して SERVING を設定すると、サーバー全体が健全とみなされる。
	h := health.NewServer()
	// デフォルト状態を SERVING に設定する
	h.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, h)

	log.Printf("server listening at:%v\n", lis.Addr())

	// Serve はブロッキングするため、メインゴルーチンを塞がないよう別ゴルーチンで実行する。
	go func() {
		if err := s.Serve(lis); err != nil {
			fmt.Printf("Failed to serve: %v", err)
		}
	}()

	// Name サービスへ「論理名 myecho → 実アドレス localhost:<port>」を登録する。
	// クライアント側のカスタム resolver が GetAddress でこの対応を引けるようにするため。
	nameServer := server.NewNameServer("localhost:60051")
	serviceName := "myecho"
	addr := fmt.Sprintf("localhost:%d", *port)
	go func() {
		nameServer.RegisterName(serviceName, addr)
	}()

	// Ctrl+C 等で終了シグナルを受け取るまでブロックする。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	<-ctx.Done()

}
