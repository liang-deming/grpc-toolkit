// gRPC サーバー側の HTTP/2 keepalive（ping フレーム）に関する設定。
// クライアントの ping 間隔が EnforcementPolicy の MinTime 未満の場合、接続が閉じられることがある。
package server

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// GetKeepaliveOpt は EnforcementPolicy と ServerParameters を束ねた ServerOption のスライスを返す。
//
// EnforcementPolicy:
//   - MinTime: クライアントが連続 ping してよい最短間隔（これより短い ping は違反）
//   - PermitWithoutStream: アクティブな RPC がなくても ping を許可するか
//
// ServerParameters:
//   - MaxConnectionIdle: アイドル状態が続いた接続を閉じるまでの時間
//   - MaxConnectionAge / Grace: 接続の最大寿命と強制終了前の猶予
//   - Time / Timeout: サーバーがアイドル接続に ping を送る間隔と、その応答待ち時間
func GetKeepaliveOpt() (opts []grpc.ServerOption) {
	// サーバー側の軽量 keepalive 方針。違反するクライアントは切断される
	var kaep = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // クライアントのアイドルタイムアウト
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		// クライアントが5秒アイドルのとき ping で保活する
		Time: 5 * time.Second,
		// ping のタイムアウト時間
		Timeout: 1 * time.Second,
	}

	return []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp)}
}
