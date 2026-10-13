package cachekv_test

import (
	"testing"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
	dbm "github.com/cosmos/cosmos-db"
)

// A store that is branched and written back without ever being touched, which
// is what most of the ~40 per-tx branches in CheckTx are.
func BenchmarkCacheKVStoreBranchUntouched(b *testing.B) {
	parent := dbadapter.Store{DB: dbm.NewMemDB()}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		st := cachekv.NewStore(parent)
		st.Write()
	}
}

func BenchmarkCacheKVStoreBranchOneWrite(b *testing.B) {
	parent := dbadapter.Store{DB: dbm.NewMemDB()}
	key, value := []byte("k"), []byte("v")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		st := cachekv.NewStore(parent)
		st.Set(key, value)
		st.Write()
	}
}
