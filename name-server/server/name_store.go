// サービス名からエンドポイントアドレスへのインメモリレジストリ。
// キーは serviceName、値は address 文字列をキーとした map（同一サービスに複数アドレスを登録可能）。
// 並行アクセスは RWMutex で保護する。
package server

import (
	"sync"
)

// Address は 1 つの登録エントリ（現状はデバッグ用フィールドが多い）。
type Address struct {
	serviceName string
	addr        string
}

// nameStore はグローバルな serviceNameData が指す実体。
type nameStore struct {
	data       map[string]map[string]*Address
	dataLocker sync.RWMutex
}

var serviceNameData *nameStore

func init() {
	serviceNameData = &nameStore{
		data: map[string]map[string]*Address{},
	}

}

// Register は serviceName に対し address を 1 件追加する。同一 address の重複は map により上書きと同様の動作。
func Register(serviceName, address string) {
	ns := serviceNameData
	addr := &Address{
		serviceName: serviceName,
		addr:        address,
	}

	ns.dataLocker.Lock()
	_, ok := ns.data[serviceName]
	if !ok {
		ns.data[serviceName] = make(map[string]*Address, 0)
	}
	ns.data[serviceName][address] = addr
	ns.dataLocker.Unlock()

}

// GetAllData はデバッグ用にストア全体へのポインタを返す。本番では公開しない方がよい。
func GetAllData() *nameStore {
	return serviceNameData
}

// GetByServiceName はサービス名に基づきアドレス情報を取得する。
// 登録された全アドレスの文字列スライスを返す。未登録なら空スライス。
func GetByServiceName(serviceName string) []string {
	ns := serviceNameData
	ns.dataLocker.RLock()
	defer ns.dataLocker.RUnlock()
	if _, ok := ns.data[serviceName]; ok {
		address := make([]string, 0)
		for _, mapv := range ns.data[serviceName] {
			address = append(address, mapv.addr)
		}
		return address
	}

	return []string{}
}
