// TLS / mTLS 用の gRPC サーバー資格情報を構築する。
// 証明書はリポジトリ内の x509 ディレクトリを参照する想定。
package server

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GetTlsOpt はサーバー証明書と秘密鍵のみを用いた TLS（クライアント認証なし）の ServerOption を返す。
// クライアントは CA 信頼のみで接続する典型的なパターン。
func GetTlsOpt() grpc.ServerOption {
	creds, err := credentials.NewServerTLSFromFile("x509/server_cert.pem", "x509/server_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	return grpc.Creds(creds)

}

// GetMTlsOpt は相互 TLS（mTLS）用の ServerOption を返す。
//
// サーバー側: server_cert / server_key で自身を提示。
// クライアント側: client_ca_cert.pem に含まれる CA で署名されたクライアント証明書を要求し検証する（RequireAndVerifyClientCert）。
// 平文接続や不正なクライアント証明書は接続段階で拒否される。
func GetMTlsOpt() grpc.ServerOption {
	cert, err := tls.LoadX509KeyPair("x509/server_cert.pem", "x509/server_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	ca := x509.NewCertPool()
	caFilePath := "x509/client_ca_cert.pem"
	bytes, err := os.ReadFile(caFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if ok := ca.AppendCertsFromPEM(bytes); !ok {
		log.Fatal("ca append failed")
	}
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
	}
	return grpc.Creds(credentials.NewTLS(tlsConfig))
}
