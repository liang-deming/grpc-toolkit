// サーバー側 gRPC インターセプタ。メタデータの Authorization ヘッダを検証する簡易 OAuth2 風の処理。
// ヘルスチェック RPC はロードバランサ等から認証なしで呼ばれることが多いため除外している。
package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryInterceptor は Unary RPC の前後にフックする。
// FullMethod が gRPC 標準ヘルスチェック以外の場合のみ oauth2Valid を実行する。
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	fmt.Println("Server UnaryInterceptor")
	fmt.Println(info)

	if info.FullMethod != "/grpc.health.v1.Health/Check" {
		err = oauth2Valid(ctx)
		if err != nil {
			return nil, err
		}
	}
	return handler(ctx, req)
}

// StreamInterceptor はストリーミング RPC 用。コンテキストから同様に認証を行う。
func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	fmt.Println("Server StreamInterceptor")
	fmt.Println(info)
	err := oauth2Valid(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, ss)
}

// oauth2Valid は gRPC メタデータから authorization キーを取り出し、Bearer トークンが期待値と一致するか検証する。
func oauth2Valid(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("メタデータの取得に失敗しました。認証に失敗しました")
	}
	authorization := md["authorization"]
	if !valid(authorization) {
		return errors.New("アクセストークンの検証に失敗しました。認証に失敗しました")
	}

	return nil
}

// valid は authorization スライスの先頭要素から "Bearer " プレフィックスを除いた値が fetchToken と一致するか判定する。
func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == fetchToken()
}

// fetchToken はサーバーが許可するアクセストークン文字列（デモ用の固定値）。
func fetchToken() string {
	return "some-secret-token"
}
