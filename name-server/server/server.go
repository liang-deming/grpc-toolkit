// name-server の gRPC ハンドラ実装。proto の Name サービスに対応する。
// Delete / Keepalive は UnimplementedNameServer に委譲されるか、未実装のままになる（proto 定義に依存）。
package server

import (
	"context"
	"fmt"
	"grpcv2/name"
	"log"
)

// NameServer は name.NameServer インターフェースの実装。
type NameServer struct {
	name.UnimplementedNameServer
}

// Register はクライアントから送られたサービス名とアドレス一覧を name_store に登録する。
// 各 address に対して Register を呼び、登録後の一覧をログに出す。
func (NameServer) Register(ctx context.Context, in *name.NameRequest) (*name.NameResponse, error) {
	for _, address := range in.Address {
		Register(in.ServiceName, address)
	}
	log.Println(GetByServiceName(in.ServiceName))
	return &name.NameResponse{ServiceName: in.ServiceName}, nil
}

// GetAddress はサービス名に紐づく実アドレス文字列のスライスを返す。
// クライアント側カスタム resolver が名前解決に利用する。
func (NameServer) GetAddress(ctx context.Context, in *name.NameRequest) (*name.NameResponse, error) {
	addr := GetByServiceName(in.ServiceName)
	fmt.Println(in.ServiceName)
	log.Println(addr)
	return &name.NameResponse{ServiceName: in.ServiceName, Address: addr}, nil
}
