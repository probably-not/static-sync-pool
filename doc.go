// Package staticsyncpool provides a wrapper around the Go Standard Library sync.Pool primitive,
// which ensures ensures that we keep a static amount of values in memory (unaffected by GC),
// while allowing us the full benefits of the sync.Pool.
package staticsyncpool
