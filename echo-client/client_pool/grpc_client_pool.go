// client_pool は grpc.ClientConn を sync.Pool で再利用する薄いラッパ。
// 取得時に接続状態が Shutdown / TransientFailure の場合は新規 Dial し直す簡易ロジックを含む。
//
// 注意: sync.Pool は GC 時にオブジェクトを捨てるため、厳密な「接続プール」としては
// 本番向けの接続管理ライブラリの方が適切な場合がある。
package client_pool

import (
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// ClientPool は Get / Put で *grpc.ClientConn を貸し借りするインターフェース。
type ClientPool interface {
	Get() *grpc.ClientConn
	Put(conn *grpc.ClientConn)
}

// clientPool は sync.Pool を内包した実装。
type clientPool struct {
	pool sync.Pool
}

// GetPool は target と DialOption を束ねたプールを作成する。
// Pool の New 関数内で grpc.Dial が呼ばれ、失敗時は nil が返る可能性がある（Get で panic の可能性）。
func GetPool(target string, opts ...grpc.DialOption) (ClientPool, error) {
	return &clientPool{
		pool: sync.Pool{
			New: func() any {
				conn, err := grpc.Dial(target, opts...)
				if err != nil {
					log.Println(err)
					return nil
				}
				return conn

			},
		},
	}, nil
}

// Get はプールから接続を取得する。状態が悪い場合は Close して新規生成。
func (c *clientPool) Get() *grpc.ClientConn {
	conn := c.pool.Get().(*grpc.ClientConn)
	if conn.GetState() == connectivity.Shutdown || conn.GetState() == connectivity.TransientFailure {
		conn.Close()
		conn = c.pool.New().(*grpc.ClientConn)
	}
	return conn

}

// Put は接続をプールに返す。既に切断済みなら Close のみ行いプールに戻さない。
func (c *clientPool) Put(conn *grpc.ClientConn) {
	if conn.GetState() == connectivity.Shutdown || conn.GetState() == connectivity.TransientFailure {
		conn.Close()
		return
	}
	c.pool.Put(conn)

}
