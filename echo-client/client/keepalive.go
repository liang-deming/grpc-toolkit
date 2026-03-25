// gRPC クライアント側の keepalive パラメータ。
// サーバー側の EnforcementPolicy（特に MinTime）と矛盾しないよう間隔を調整することが重要。
package client

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// GetKeepaliveOpt は grpc.WithKeepaliveParams に渡す ClientParameters を構築する。
// Time: ping 送信間隔、Timeout: ping に対する ACK の待ち時間、PermitWithoutStream: RPC が無いときも ping するか。
func GetKeepaliveOpt() (opt grpc.DialOption) {
	var kacp = keepalive.ClientParameters{
		// アクティブなストリームがない場合、10秒ごとに ping を送る
		Time: 10 * time.Second,
		// ping のタイムアウト時間
		Timeout: time.Second,
		// アクティブなストリームがない場合でも ping を受け入れるか
		PermitWithoutStream: true,
	}
	return grpc.WithKeepaliveParams(kacp)
}
