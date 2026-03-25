// OAuth2 トークンを gRPC の PerRPCCredentials として渡すためのヘルパ。
// golang.org/x/oauth2 の StaticTokenSource と grpc の oauth.TokenSource を組み合わせている。
package client

import (
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// GetAuth は Dial 時に WithPerRPCCredentials を付与するための DialOption。
// すべての RPC で同一トークンが Authorization に載る（サーバー側 interceptor と一致させる）。
func GetAuth(token string) grpc.DialOption {
	perRPC := GetPerRPCCredentials(token)
	return grpc.WithPerRPCCredentials(perRPC)
}

// GetPerRPCCredentials は OAuth2 アクセストークンを gRPC がメタデータに載せる形にラップする。
func GetPerRPCCredentials(token string) credentials.PerRPCCredentials {
	return oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})}
}

// FetchToken はデモ用の固定トークン。サーバー側 fetchToken と同じ値である必要がある。
func FetchToken() string {
	return "some-secret-token"
}
