package staticsyncpool

type config struct {
	staticSize int
	lazy       bool
}

func defaultConfig() *config {
	return &config{
		staticSize: 100,
		lazy:       false,
	}
}

type Option interface {
	apply(*config)
}

var _ Option = (*option)(nil)

type option struct {
	f func(*config)
}

func newOption(f func(*config)) *option {
	return &option{
		f: f,
	}
}

func (o *option) apply(c *config) {
	o.f(c)
}

// WithStaticSize configures the static size of the pool.
// The static size defines the minimum number of items within the pool at any given time.
// A higher number means that the pool will hold a higher number in memory, thus incurring a runtime memory cost.
// However, a lower number means that the pool will hold less static items in memory, and will thus incur a runtime GC cost.
// This can be used to tune performance based on the expected size of the pool, and whether the pool is in a hot-path in the code.
func WithStaticSize(staticSize int) Option {
	return newOption(func(c *config) {
		c.staticSize = staticSize
	})
}

// WithLazy configures the lazy setting of the pool.
// The lazy setting defines whether the pool will lazily initialize a pool of the static size,
// or whether the initialization will be eager.
// If Lazy is set to false, the pool will eagerly create enough items to fill the pool at the required minimum static size.
// This incurs a startup cost, but will mean that the runtime cost of `Get`-ing a value from the pool will be lower, since
// the pool will have a minimum static size.
// If Lazy is set to true, the pool will lazily create enough items to fill the pool at the required minimum static size.
// This incurs a potential runtime cost, as the calls to `Get` will all allocate (unless the pool has values in it already).
// This value can be used to tune performance:
// For example: if the pool is in a hot-path in the code, then it may be beneficial to use
// `WithLazy(false)`, and eagerly create the minimum pool size on startup, thus lowering the runtime cost.
// However, if the pool is not in a hot-path of the code, then it may be beneficial to use `WithLazy(true)`, since the pool will have
// more hits in re-using the values within it.
func WithLazy(lazy bool) Option {
	return newOption(func(c *config) {
		c.lazy = lazy
	})
}
