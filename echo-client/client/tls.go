// クライアント側 TLS / mTLS 用の TransportCredentials を構築する。
// サーバー側 echo-server/server/tls.go の設定と証明書（x509 配下）と対になる。
package client

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GetTlsOpt はサーバー証明書を CA で検証する TLS のみ（クライアント証明書なし）の DialOption。
// ServerName は証明書の SAN / CN と一致させる必要がある（echo.grpc.0voice.com）。
func GetTlsOpt() grpc.DialOption {
	creds, err := credentials.NewClientTLSFromFile("x509/ca_cert.pem", "echo.grpc.0voice.com")
	if err != nil {
		log.Fatal(err)
	}
	opt := grpc.WithTransportCredentials(creds)
	return opt

}

// GetMTlsOpt はクライアント証明書を提示し、サーバー側でクライアント認証する mTLS 用 DialOption。
// RootCAs でサーバー証明書の検証、Certificates で自身を提示する。
// ServerName は実際の証明書と一致させること（例では abc.grpc.0voice.com）。
func GetMTlsOpt() grpc.DialOption {
	cert, err := tls.LoadX509KeyPair("x509/client_cert.pem", "x509/client_key.pem")
	if err != nil {
		log.Fatal(err)
	}
	ca := x509.NewCertPool()
	caFilePath := "x509/ca_cert.pem"
	bytes, err := os.ReadFile(caFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if ok := ca.AppendCertsFromPEM(bytes); !ok {
		log.Fatal("ca append failed")
	}
	tlsConfig := &tls.Config{
		ServerName:   "abc.grpc.0voice.com",
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
	}
	return grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
}
