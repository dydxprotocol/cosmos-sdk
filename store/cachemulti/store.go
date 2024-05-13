package cachemulti

import (
	"fmt"
	"io"
	"sync"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"golang.org/x/exp/slices"
)

// storeNameCtxKey is the TraceContext metadata key that identifies
// the store which emitted a given trace.
const storeNameCtxKey = "store_name"

//----------------------------------------
// Store

// Store holds many branched stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	db     types.CacheKVStore
	stores map[types.StoreKey]types.CacheWrap
	keys   map[string]types.StoreKey

	traceWriter  io.Writer
	traceContext types.TraceContext

	locks *sync.Map // map from string key to *sync.Mutex or *sync.RWMutex
}

var (
	_ types.CacheMultiStore = Store{}
)

// NewFromKVStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects and a KVStore as the database. Each CacheWrapper store
// is a branched store.
func NewFromKVStore(
	store types.KVStore,
	stores map[types.StoreKey]types.CacheWrapper,
	keys map[string]types.StoreKey,
	traceWriter io.Writer,
	traceContext types.TraceContext,
	locks *sync.Map,
) Store {
	cms := Store{
		db:           cachekv.NewStore(store),
		stores:       make(map[types.StoreKey]types.CacheWrap, len(stores)),
		keys:         keys,
		traceWriter:  traceWriter,
		traceContext: traceContext,
		locks:        locks,
	}

	for key, store := range stores {
		if cms.TracingEnabled() {
			tctx := cms.traceContext.Clone().Merge(types.TraceContext{
				storeNameCtxKey: key.Name(),
			})

			store = tracekv.NewStore(store.(types.KVStore), cms.traceWriter, tctx)
		}
		cms.stores[key] = cachekv.NewStore(store.(types.KVStore))
	}

	return cms
}

// NewStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects. Each CacheWrapper store is a branched store.
func NewStore(
	db dbm.DB, stores map[types.StoreKey]types.CacheWrapper, keys map[string]types.StoreKey,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	return NewFromKVStore(dbadapter.Store{DB: db}, stores, keys, traceWriter, traceContext, nil)
}

// NewLockingStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects. Each CacheWrapper store is a branched store.
func NewLockingStore(
	db dbm.DB, stores map[types.StoreKey]types.CacheWrapper, keys map[string]types.StoreKey,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	return NewFromKVStore(
		dbadapter.Store{DB: db},
		stores,
		keys,
		traceWriter,
		traceContext,
		&sync.Map{},
	)
}

func newCacheMultiStoreFromCMS(cms Store) Store {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range cms.stores {
		stores[k] = v
	}

	return NewFromKVStore(cms.db, stores, nil, cms.traceWriter, cms.traceContext, cms.locks)
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms Store) SetTracer(w io.Writer) types.MultiStore {
	cms.traceWriter = w
	return cms
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms Store) SetTracingContext(tc types.TraceContext) types.MultiStore {
	if cms.traceContext != nil {
		for k, v := range tc {
			cms.traceContext[k] = v
		}
	} else {
		cms.traceContext = tc
	}

	return cms
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (cms Store) TracingEnabled() bool {
	return cms.traceWriter != nil
}

// LatestVersion returns the branch version of the store
func (cms Store) LatestVersion() int64 {
	panic("cannot get latest version from branch cached multi-store")
}

// GetStoreType returns the type of the store.
func (cms Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Write calls Write on each underlying store.
func (cms Store) Write() {
	cms.db.Write()
	for _, store := range cms.stores {
		store.Write()
	}
}

// Lock, Unlock, RLockRW, LockRW, RUnlockRW, UnlockRW constitute a permissive locking interface
// that can be used to synchronize concurrent access to the store. Locking of a key should
// represent locking of some part of the store. Note that improper access is not enforced, and it is
// the user's responsibility to ensure proper locking of any access by concurrent goroutines.
//
// Common mistakes may include:
//   - Introducing data races by reading or writing state that is claimed by a competing goroutine
//   - Introducing deadlocks by locking in different orders through multiple calls to locking methods.
//     i.e. if A calls Lock(a) followed by Lock(b), and B calls Lock(b) followed by Lock(a)
//   - Using a key as an exclusive lock after it has already been initialized as a read-write lock

// Lock acquires exclusive locks on a set of keys.
func (cms Store) Lock(keys [][]byte) {
	for _, stringKey := range keysToSortedStrings(keys) {
		v, _ := cms.locks.LoadOrStore(stringKey, &sync.Mutex{})
		lock := v.(*sync.Mutex)
		lock.Lock()
	}
}

// Unlock releases exclusive locks on a set of keys.
func (cms Store) Unlock(keys [][]byte) {
	for _, key := range keys {
		v, ok := cms.locks.Load(string(key))
		if !ok {
			panic("Key not found")
		}
		lock := v.(*sync.Mutex)
		lock.Unlock()
	}
}

// RLockRW acquires read locks on a set of keys.
func (cms Store) RLockRW(keys [][]byte) {
	for _, stringKey := range keysToSortedStrings(keys) {
		v, _ := cms.locks.LoadOrStore(stringKey, &sync.RWMutex{})
		lock := v.(*sync.RWMutex)
		lock.RLock()
	}
}

// LockRW acquires write locks on a set of keys.
func (cms Store) LockRW(keys [][]byte) {
	for _, stringKey := range keysToSortedStrings(keys) {
		v, _ := cms.locks.LoadOrStore(stringKey, &sync.RWMutex{})
		lock := v.(*sync.RWMutex)
		lock.Lock()
	}
}

// RUnlockRW releases read locks on a set of keys.
func (cms Store) RUnlockRW(keys [][]byte) {
	for _, key := range keys {
		v, ok := cms.locks.Load(string(key))
		if !ok {
			panic("Key not found")
		}
		lock := v.(*sync.RWMutex)
		lock.RUnlock()
	}
}

// UnlockRW releases write locks on a set of keys.
func (cms Store) UnlockRW(keys [][]byte) {
	for _, key := range keys {
		v, ok := cms.locks.Load(string(key))
		if !ok {
			panic("Key not found")
		}
		lock := v.(*sync.RWMutex)
		lock.Unlock()
	}
}

func keysToSortedStrings(keys [][]byte) []string {
	// Ensure that we always operate in a deterministic ordering when acquiring locks to prevent deadlock.
	stringLockedKeys := make([]string, len(keys))
	for i, key := range keys {
		stringLockedKeys[i] = string(key)
	}
	slices.Sort(stringLockedKeys)
	return stringLockedKeys
}

// Implements CacheWrapper.
func (cms Store) CacheWrap() types.CacheWrap {
	return cms.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return cms.CacheWrap()
}

// Implements MultiStore.
func (cms Store) CacheMultiStore() types.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// CacheMultiStoreWithVersion implements the MultiStore interface. It will panic
// as an already cached multi-store cannot load previous versions.
//
// TODO: The store implementation can possibly be modified to support this as it
// seems safe to load previous versions (heights).
func (cms Store) CacheMultiStoreWithVersion(_ int64) (types.CacheMultiStore, error) {
	panic("cannot branch cached multi-store with a version")
}

// GetStore returns an underlying Store by key.
func (cms Store) GetStore(key types.StoreKey) types.Store {
	s := cms.stores[key]
	if key == nil || s == nil {
		panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
	}
	return s.(types.Store)
}

// GetKVStore returns an underlying KVStore by key.
func (cms Store) GetKVStore(key types.StoreKey) types.KVStore {
	store := cms.stores[key]
	if key == nil || store == nil {
		panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
	}
	return store.(types.KVStore)
}
