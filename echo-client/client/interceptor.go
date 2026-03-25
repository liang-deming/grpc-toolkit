// クライアント側 gRPC インターセプタ。Unary と Stream で PerRPC 資格情報の渡し方が異なるため、
// デバッグ用に CallOption の型（ポインタ/値）をログ出力している。
// Unary 側では PerRPCCredsCallOption のポインタ検出が期待どおりに動かない場合がある（コメント参照）。
package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// UnaryInterceptor は Unary RPC の invoker 呼び出し前に実行される。
// opts 内に grpc.PerRPCCredsCallOption が含まれるかを調べるが、grpc.PerRPCCredentials で渡した場合の
// 内部表現がポインタと一致しないことがあり、期待どおりの動作にならない点に注意。
func UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Println("client UnaryInterceptor")
	var credsConfigured bool
	for _, opt := range opts {
		_, isPtr := opt.(*grpc.PerRPCCredsCallOption)
		_, isVal := opt.(grpc.PerRPCCredsCallOption)

		fmt.Printf("Is Pointer: %v, Is Value: %v, Real Type: %T\n", isPtr, isVal, opt)
	}
	for _, opt := range opts {
		_, ok := opt.(*grpc.PerRPCCredsCallOption) // 期待どおりの動作にならない
		if ok {
			credsConfigured = true
			break
		}
	}
	if !credsConfigured {
		//opts = append(opts, grpc.PerRPCCredentials(GetPerRPCCredentials(FetchToken())))
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamInterceptor はストリーミング RPC 用。値型の PerRPCCredsCallOption を検出する。
// 資格情報が無い場合は opts に PerRPCCredentials を追加してから streamer を呼ぶ。
func StreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	fmt.Println("client StreamInterceptor")
	var credsConfigured bool
	for _, opt := range opts {
		_, isPtr := opt.(*grpc.PerRPCCredsCallOption)
		_, isVal := opt.(grpc.PerRPCCredsCallOption)

		fmt.Printf("Is Pointer: %v, Is Value: %v, Real Type: %T\n", isPtr, isVal, opt)
	}
	for _, opt := range opts {
		_, ok := opt.(grpc.PerRPCCredsCallOption)
		if ok {
			credsConfigured = true
			break
		}
	}

	if !credsConfigured {
		opts = append(opts, grpc.PerRPCCredentials(GetPerRPCCredentials(FetchToken())))
	}
	return streamer(ctx, desc, cc, method, opts...)
}
