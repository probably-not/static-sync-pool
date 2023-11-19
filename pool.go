package staticsyncpool

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Pool is a wrapper around Go's `sync.Pool` which allows configuring a static number of
// items that are not automatically Garbage Collected.
type Pool[T any] struct {
	resetMu      *sync.Mutex
	config       *config
	pinner       *runtime.Pinner
	newFunc      func() T
	resetFunc    func(T)
	internalPool *sync.Pool
	lazySize     atomic.Int64
	closed       atomic.Bool
}

// New will initialize a new Pool with the given `newFunc` for initializing new values,
// the given `resetFunc` for resetting the values before putting them back into the pool,
// and applying the given options to the config.
// The default config sets the Static Size to 100 and the Lazy setting to false.
// These can be adjusted using `WithStaticSize` and `WithLazy`.
func New[T any](newFunc func() *T, resetFunc func(*T), opts ...Option) *Pool[*T] {
	config := defaultConfig()
	for _, opt := range opts {
		opt.apply(config)
	}

	p := &Pool[*T]{
		resetMu:   &sync.Mutex{},
		config:    config,
		newFunc:   newFunc,
		resetFunc: resetFunc,
		pinner:    &runtime.Pinner{},
		lazySize:  atomic.Int64{},
		closed:    atomic.Bool{},
		internalPool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}

	p.init()

	return p
}

func (p *Pool[T]) init() {
	// Set the lazy size to 0 since this is an initialization...
	// we should have 0 values in the size
	p.lazySize.Store(0)

	if p.config.lazy {
		return
	}

	for i := 0; i < p.config.staticSize; i++ {
		item := p.newFunc()
		p.pinner.Pin(item)
		p.internalPool.Put(item)
	}

	p.lazySize.Add(int64(p.config.staticSize))
}

// Get will return an item from the Pool.
// If the pool was configured with lazy as false,
// then it does nothing else.
// If the pool was configured with lazy as true,
// then extra logic is run to determine if the pool needs to add this value to it's
// static pool. If the pool is currently closed (by the Reset function),
// then we simply return the result of the configured `newFunc`.
func (p *Pool[T]) Get() T {
	if p.closed.Load() {
		// If the pool is closed, do nothing special, just allocate.
		return p.newFunc()
	}

	item := p.internalPool.Get().(T)

	// If the config is non-lazy, we've already initialized the pool with a static amount of pinned values,
	// so no need to do checks for pinning.
	if !p.config.lazy {
		return item
	}

	currentLazySize := p.lazySize.Load()
	// If we've already initialized enough values, no need to run the pinning.
	if currentLazySize >= int64(p.config.staticSize) {
		return item
	}

	p.pinner.Pin(item)
	// Racy-ish... we can potentially create too many without doing a CAS.
	// But, doing a CAS could also be potentially heavy... depending on hot-path and everything.
	// Docs should note that when `WithLazy` is set to `true`, the pool can become semi-leaky...
	// meaning we may create more items than the static size configuration sets.
	// `WithStaticSize` defines a minimum number of static elements, however because this isn't being done with CAS
	// then it's possible to pass it.
	p.lazySize.Add(1)

	return item
}

// Put will return an item to the pool, to be reused by others.
// Before returning to the pool, the configured `resetFunc` is run
// to reset the item for reuse.
func (p *Pool[T]) Put(item T) {
	p.resetFunc(item)
	p.internalPool.Put(item)
}

// Reset will for a complete reset of the pool's memory.
// This will unpin all memory that has been added to the static pool.
// This does not release memory - this is left to Go's GC and sync.Pool implementations.
// When using Reset, the configuration is automatically set to `Lazy = true`.
// This is for practical purposes:
// If you are using `Reset`, then there should be an expectation that
// all memory is released from the static values and you want your memory usage to decrease.
// It is recommended that if you have set the lazy configuration to false that you do not reset,
// since you will lose the performance boost of having your static pool size initialized at startup.
func (p *Pool[T]) Reset() {
	// Acquire lock. Reset should only run 1 at a time. This is the only function that uses a mutext,
	// other functions utilize `p.closed` to ensure that they are running without affecting the pinner.
	p.resetMu.Lock()
	defer p.resetMu.Unlock()

	// Close the pool so that we don't have any leaks in the pinner
	p.closed.Store(true)
	// Unpin everything
	p.pinner.Unpin()
	// Force the config to be lazy after Reset completes.
	p.config.lazy = true
	// Initialize the pool from scratch
	p.init()
	// Set closed back to false to open the pool back up
	p.closed.Store(false)
}
