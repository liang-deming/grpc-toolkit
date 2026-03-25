// Name サービス（name-server）への gRPC クライアントラッパ。
// Echo サーバー起動時に、自ホストの listen アドレスを論理サービス名に紐づけて登録するために使う。
package server

import (
	"context"
	"grpcv2/name"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NameServer は name パッケージの Name サービスへ接続するための薄いヘルパ。
// conn は insecure（平文）で name-server にダイアルした ClientConn。
type NameServer struct {
	conn *grpc.ClientConn
}

// NewNameServer は指定アドレス（通常 "host:port"）の Name gRPC サーバーへ接続する。
// Dial 失敗時も recover で握りつぶさずログに出すだけで conn が nil の可能性がある点に注意。
func NewNameServer(addr string) *NameServer {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
	}
	return &NameServer{
		conn: conn,
	}
}

// RegisterName は Register RPC を呼び出し、serviceName に対して addr を 1 件登録する。
// クライアント側 resolver が GetAddress で取得できるようにするための自己登録（サービス登録）に相当する。
func (ns *NameServer) RegisterName(serviceName, addr string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	client := name.NewNameClient(ns.conn)
	in := &name.NameRequest{
		ServiceName: serviceName,
		Address:     []string{addr},
	}
	_, err := client.Register(context.Background(), in)
	if err != nil {
		log.Println(err)
	}

}
