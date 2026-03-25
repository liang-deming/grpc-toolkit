// name-server は Name gRPC サービス（サービス名→アドレス一覧のレジストリ）のエントリポイント。
// インメモリの name_store に対し、Register / GetAddress 等の RPC を処理する。
// Echo サーバーが起動時に Register し、Echo クライアントの resolver が GetAddress で取得する、という流れに使う。
package main

import (
	"flag"
	"fmt"
	"grpcv2/name"
	"grpcv2/name-server/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 60051, "")
)

func main() {
	//testdata()

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	name.RegisterNameServer(s, &server.NameServer{})
	log.Printf("server listening at : %v", lis.Addr())
	err = s.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}

}

// testdata は手動で store に登録し、GetByServiceName の結果を確認するデバッグ用関数。
func testdata() {
	// 最終的に NameServer のメモリ上に次のようなデータができる：
	server.Register("myecho", "localhost:50051")
	alldata := server.GetAllData()
	fmt.Println(alldata)
	fmt.Println(server.GetByServiceName("myecho"))
}
