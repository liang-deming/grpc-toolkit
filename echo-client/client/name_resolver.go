// gRPC のカスタム name resolver 実装。
// target が "myscheme://myecho" のような形式のとき、エンドポイント名をキーに Name サービスから
// 実アドレス一覧を取得し、resolver.State に載せて gRPC 接続レイヤへ渡す。
//
// 注意: ResolveNow はデモ用に固定アドレスへ差し替えており、本番運用向けではない。
package client

import (
	"context"
	"grpcv2/name"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

const (
	// MyScheme は grpc.Dial の target で使用するカスタムスキーム名。RegisterWithResolvers される Builder と一致させる。
	MyScheme = "myscheme"
	// MyServiceName は Name サービス上の論理サービス名。Echo サーバーが登録するときと同じ文字列にする。
	MyServiceName = "myecho"
)

// var addrs = []string{"localhost:50051", "localhost:50052", "localhost:50053"}
var nameServer *NameServer

// GetNameResolver は grpc.WithResolvers に渡す DialOption を返す。
// グローバル nameServer を設定するため、複数の異なる Name 接続を並行で使う用途には向かない。
func GetNameResolver(ns *NameServer) grpc.DialOption {
	nameServer = ns
	return grpc.WithResolvers(&MyResolverBuilder{})
}

// MyResolverBuilder は resolver.Builder 実装。Scheme() が MyScheme を返す。
type MyResolverBuilder struct {
}

// Build は新しい MyResolver を生成し、初回の名前解決結果を Name サービスから取得して State を更新する。
func (*MyResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &MyResolver{
		target:     target,
		cc:         cc,
		addrsStore: map[string][]string{MyServiceName: nameServer.getAddressByServiceName(MyServiceName)},
	}
	r.start()
	return r, nil
}

// Scheme はこの Builder が扱う URL スキーム（"myscheme"）を返す。
func (*MyResolverBuilder) Scheme() string {
	return MyScheme
}

// MyResolver は単一の resolver.Target に対する解決状態を保持する。
type MyResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

// start は addrsStore からエンドポイントに対応するアドレス文字列を取り出し、
// resolver.Address スライスに変換して ClientConn.UpdateState を呼ぶ。
func (r *MyResolver) start() {
	log.Println("Resolver start")
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{
			Addr: s,
		}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow は名前解決の再実行が要求されたときに呼ばれる。
// ここではデモ用に固定の 3 アドレスへ差し替えており、本番の動的更新の例としては不適切。
func (r *MyResolver) ResolveNow(o resolver.ResolveNowOptions) {
	log.Println("Resolve Now")
	log.Println(r.cc)
	r.addrsStore = map[string][]string{MyServiceName: {"localhost:50054", "localhost:50055", "localhost:50056"}}
	r.start()
	log.Println(r.cc)
}

// Close はリソース解放用。現状は何もしない。
func (r *MyResolver) Close() {}

// NameServer は name パッケージの gRPC サービスへ接続するクライアント側の薄いラッパ。
type NameServer struct {
	conn *grpc.ClientConn
}

// NewNameServer は Name サービス（通常 name-server）へ insecure で接続する。
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

// getAddressByServiceName は GetAddress RPC を呼び、登録済みアドレスのスライスを返す。
// エラー時は空スライスとログ出力。
func (ns *NameServer) getAddressByServiceName(serviceName string) []string {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	client := name.NewNameClient(ns.conn)
	in := &name.NameRequest{
		ServiceName: serviceName,
	}
	res, err := client.GetAddress(context.Background(), in)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	log.Println(res.Address)
	return res.Address
}
