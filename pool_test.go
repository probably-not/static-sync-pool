package staticsyncpool

import (
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Example struct {
	A string
	B int64
	C float64
}

func TestPoolALotToSeeIfAnythingHappens(t *testing.T) {
	for i := 0; i < 1000; i++ {
		testPool(t, i)
	}
}

func testPool(t *testing.T, iteration int) {
	const staticSize = 100
	pool := New[Example](
		func() *Example {
			return &Example{}
		},
		func(es *Example) {
			es.A = "I've been reset"
			es.B = 0
			es.C = 0.0
		},
		WithLazy(false),
		WithStaticSize(staticSize),
	)

	items := make([]*Example, 0, staticSize)
	for i := 0; i < staticSize; i++ {
		fromPool := pool.Get()
		diff := cmp.Diff(fromPool, &Example{})
		if diff != "" {
			t.Errorf("test iteration %d; expected fromPool to be empty during initial loop; iteration %d; diff from empty=\n%s", iteration, i, diff)
			return
		}
		items = append(items, fromPool)
	}

	// Return to pool
	for _, item := range items {
		pool.Put(item)
	}

	for i := 0; i < staticSize; i++ {
		fromPool := pool.Get()
		diff := cmp.Diff(fromPool, &Example{
			A: "I've been reset",
		})
		if diff != "" {
			t.Errorf("test iteration %d; expected fromPool to have been reset on second loop; iteration %d; diff from expected=\n%s", iteration, i, diff)
			return
		}
	}

	fromPool := pool.Get()
	diff := cmp.Diff(fromPool, &Example{})
	if diff != "" {
		t.Errorf("test iteration %d; expected fromPool to be empty after all static values are extracted from pool in secondary loop; diff from empty=\n%s", iteration, diff)
		return
	}
	pool.Put(fromPool)

	pool.Reset()
	// IIRC sync.Pool has a few GCs before it actually collects the values.
	runtime.GC()
	runtime.GC()
	runtime.GC()

	fromPool = pool.Get()
	diff = cmp.Diff(fromPool, &Example{})
	if diff != "" {
		t.Errorf("test iteration %d; expected fromPool to be empty after Reset() and GC run; diff from empty=\n%s", iteration, diff)
		return
	}

	items = make([]*Example, 0, staticSize)
	items = append(items, fromPool)
	for i := 0; i < staticSize-1; /* subtract 1 because we did 1 above in the check */ i++ {
		fromPool := pool.Get()
		diff := cmp.Diff(fromPool, &Example{})
		if diff != "" {
			t.Errorf("test iteration %d; expected fromPool to be empty during initial loop on lazy iteration; iteration %d; diff from empty=\n%s", iteration, i, diff)
			return
		}
		items = append(items, fromPool)
	}

	// Return to pool
	for _, item := range items {
		pool.Put(item)
	}

	for i := 0; i < staticSize; i++ {
		fromPool := pool.Get()
		diff := cmp.Diff(fromPool, &Example{
			A: "I've been reset",
		})
		if diff != "" {
			t.Errorf("test iteration %d; expected fromPool to have been reset on second loop on lazy iteration; iteration %d; diff from expected=\n%s", iteration, i, diff)
			return
		}
	}

	pool.Reset()
}
