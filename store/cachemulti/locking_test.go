package cachemulti_test

import (
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_LinearizeReadsAndWrites(t *testing.T) {
	key := []byte("kv_store_key")
	storeKey := types.NewKVStoreKey("store1")
	lockKey := []byte("a")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			lockingCms.Lock([][]byte{lockKey})
			defer lockingCms.Unlock([][]byte{lockKey})
			kvStore := lockingCms.GetKVStore(storeKey)
			v := kvStore.Get(key)
			if v == nil {
				kvStore.Set(key, []byte{1})
			} else {
				v[0]++
				kvStore.Set(key, v)
			}
			lockingCms.Write()
		}()
	}

	wg.Wait()
	require.Equal(t, []byte{100}, lockingCms.GetKVStore(storeKey).Get(key))
}

func TestStore_LockOrderToPreventDeadlock(t *testing.T) {
	key := []byte("kv_store_key")
	storeKey := types.NewKVStoreKey("store1")
	lockKeyA := []byte("a")
	lockKeyB := []byte("b")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	// Acquire keys in two different orders ensuring that we don't reach deadlock.
	wg := sync.WaitGroup{}
	wg.Add(200)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			lockingCms.Lock([][]byte{lockKeyA, lockKeyB})
			defer lockingCms.Unlock([][]byte{lockKeyA, lockKeyB})
			kvStore := lockingCms.GetKVStore(storeKey)
			v := kvStore.Get(key)
			if v == nil {
				kvStore.Set(key, []byte{1})
			} else {
				v[0]++
				kvStore.Set(key, v)
			}
			lockingCms.Write()
		}()

		go func() {
			defer wg.Done()

			lockingCms.Lock([][]byte{lockKeyB, lockKeyA})
			defer lockingCms.Unlock([][]byte{lockKeyB, lockKeyA})
			kvStore := lockingCms.GetKVStore(storeKey)
			v := kvStore.Get(key)
			if v == nil {
				kvStore.Set(key, []byte{1})
			} else {
				v[0]++
				kvStore.Set(key, v)
			}
			lockingCms.Write()
		}()
	}

	wg.Wait()
	require.Equal(t, []byte{200}, lockingCms.GetKVStore(storeKey).Get(key))
}

func TestStore_AllowForParallelUpdates(t *testing.T) {
	storeKey := types.NewKVStoreKey("store1")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := byte(0); i < 100; i++ {
		k := []byte{i}
		go func() {
			defer wg.Done()

			// We specifically don't unlock the keys during processing so that we can show that we must process all
			// of these in parallel before the wait group is done.
			lockingCms.Lock([][]byte{k})
			lockingCms.GetKVStore(storeKey).Set(k, k)
			lockingCms.Write()
		}()
	}

	wg.Wait()
	for i := byte(0); i < 100; i++ {
		lockingCms.Unlock([][]byte{{i}})
	}
	for i := byte(0); i < 100; i++ {
		require.Equal(t, []byte{i}, lockingCms.GetKVStore(storeKey).Get([]byte{i}))
	}
}

func TestStore_AddLocksDuringTransaction(t *testing.T) {
	key := []byte("kv_store_key")
	storeKey := types.NewKVStoreKey("store1")
	lockKey := []byte("lockkey")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := byte(0); i < 100; i++ {
		k := []byte{i}
		go func() {
			defer wg.Done()

			lockingCms.Lock([][]byte{k})
			defer lockingCms.Unlock([][]byte{k})
			lockingCms.GetKVStore(storeKey).Set(k, k)

			lockingCms.Lock([][]byte{lockKey})
			defer lockingCms.Unlock([][]byte{lockKey})
			kvStore := lockingCms.GetKVStore(storeKey)
			v := kvStore.Get(key)
			if v == nil {
				kvStore.Set(key, []byte{1})
			} else {
				v[0]++
				kvStore.Set(key, v)
			}
			lockingCms.Write()
		}()
	}

	wg.Wait()
	for i := byte(0); i < 100; i++ {
		require.Equal(t, []byte{i}, lockingCms.GetKVStore(storeKey).Get([]byte{i}))
	}
	require.Equal(t, []byte{100}, lockingCms.GetKVStore(storeKey).Get(key))
}

func TestStore_MaintainLockOverMultipleTransactions(t *testing.T) {
	keyA := []byte("kv_store_key_a")
	keyB := []byte("kv_store_key_b")
	storeKey := types.NewKVStoreKey("store1")
	lockKeyA := []byte("lockkeya")
	lockKeyB := []byte("lockkeyb")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	// Key A is set differently in the first and second transaction so we can check it
	// to see what transaction was run last.
	lockingCms.GetKVStore(storeKey).Set(keyA, []byte{0})
	lockingCms.GetKVStore(storeKey).Set(keyB, []byte{0})

	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := byte(0); i < 100; i++ {
		k := []byte{i}
		go func() {
			defer wg.Done()

			lockingCms.Lock([][]byte{k})
			defer lockingCms.Unlock([][]byte{k})
			lockingCms.GetKVStore(storeKey).Set(k, k)

			lockingCms.Lock([][]byte{lockKeyA})
			defer lockingCms.Unlock([][]byte{lockKeyA})

			func() {
				lockingCms.Lock([][]byte{lockKeyB})
				defer lockingCms.Unlock([][]byte{lockKeyB})

				assert.Equal(t, []byte{0}, lockingCms.GetKVStore(storeKey).Get(keyA))
				lockingCms.GetKVStore(storeKey).Set(keyA, []byte{1})
				v := lockingCms.GetKVStore(storeKey).Get(keyB)
				v[0]++
				lockingCms.GetKVStore(storeKey).Set(keyB, v)
				lockingCms.Write()
			}()

			func() {
				lockingCms.Lock([][]byte{lockKeyB})
				defer lockingCms.Unlock([][]byte{lockKeyB})

				assert.Equal(t, []byte{1}, lockingCms.GetKVStore(storeKey).Get(keyA))
				lockingCms.GetKVStore(storeKey).Set(keyA, []byte{0})
				v := lockingCms.GetKVStore(storeKey).Get(keyB)
				v[0]++
				lockingCms.GetKVStore(storeKey).Set(keyB, v)
				lockingCms.Write()
			}()
		}()
	}

	wg.Wait()
	require.Equal(t, []byte{200}, lockingCms.GetKVStore(storeKey).Get(keyB))
}

func TestStore_ReadWriteLock(t *testing.T) {
	numReadersKey := []byte("kv_store_key_a")
	numWritersKey := []byte("kv_store_key_b")
	maxNumReadersKey := []byte("kv_store_key_c")
	maxNumWritersKey := []byte("kv_store_key_d")
	storeKey := types.NewKVStoreKey("store1")
	rwLockKey := []byte("lockkeya")
	lockKey := []byte("lockkeyb")

	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	store.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	store.MountStoreWithDB(storeKey, types.StoreTypeIAVL, db)
	err := store.LoadLatestVersion()
	assert.NoError(t, err)
	lockingCms := store.LockingCacheMultiStore()

	lockingCms.GetKVStore(storeKey).Set(numReadersKey, []byte{0})
	lockingCms.GetKVStore(storeKey).Set(numWritersKey, []byte{0})
	lockingCms.GetKVStore(storeKey).Set(maxNumReadersKey, []byte{0})
	lockingCms.GetKVStore(storeKey).Set(maxNumWritersKey, []byte{0})

	wg := sync.WaitGroup{}
	wg.Add(200)
	// Start 100 readers and 100 writers. Record the maximum number of readers and writers seen.
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			lockingCms.RLockRW([][]byte{rwLockKey})
			defer lockingCms.RUnlockRW([][]byte{rwLockKey})

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				v := lockingCms.GetKVStore(storeKey).Get(numReadersKey)
				v[0]++
				lockingCms.GetKVStore(storeKey).Set(numReadersKey, v)
				lockingCms.Write()
			}()

			time.Sleep(100 * time.Millisecond)

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				numReaders := lockingCms.GetKVStore(storeKey).Get(numReadersKey)[0]
				maxNumReaders := lockingCms.GetKVStore(storeKey).Get(maxNumReadersKey)[0]
				if numReaders > maxNumReaders {
					lockingCms.GetKVStore(storeKey).Set(maxNumReadersKey, []byte{numReaders})
				}
				lockingCms.Write()
			}()

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				v := lockingCms.GetKVStore(storeKey).Get(numReadersKey)
				v[0]--
				lockingCms.GetKVStore(storeKey).Set(numReadersKey, v)
				lockingCms.Write()
			}()
		}()

		go func() {
			defer wg.Done()

			lockingCms.LockRW([][]byte{rwLockKey})
			defer lockingCms.UnlockRW([][]byte{rwLockKey})

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				v := lockingCms.GetKVStore(storeKey).Get(numWritersKey)
				v[0]++
				lockingCms.GetKVStore(storeKey).Set(numWritersKey, v)
				lockingCms.Write()
			}()

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				numWriters := lockingCms.GetKVStore(storeKey).Get(numWritersKey)[0]
				maxNumWriters := lockingCms.GetKVStore(storeKey).Get(maxNumWritersKey)[0]
				if numWriters > maxNumWriters {
					lockingCms.GetKVStore(storeKey).Set(maxNumWritersKey, []byte{numWriters})
				}
				lockingCms.Write()
				lockingCms.Write()
			}()

			func() {
				lockingCms.Lock([][]byte{lockKey})
				defer lockingCms.Unlock([][]byte{lockKey})
				v := lockingCms.GetKVStore(storeKey).Get(numWritersKey)
				v[0]--
				lockingCms.GetKVStore(storeKey).Set(numWritersKey, v)
				lockingCms.Write()
			}()
		}()
	}

	wg.Wait()
	// At some point there should be more than one reader. If this test is flaky, sleep time
	// can be added to the reader to deflake.
	require.Less(t, []byte{1}, lockingCms.GetKVStore(storeKey).Get(maxNumReadersKey))
	// There must be at most one writer at once.
	require.Equal(t, []byte{1}, lockingCms.GetKVStore(storeKey).Get(maxNumWritersKey))
}
