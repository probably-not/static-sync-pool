# static-sync-pool

A wrapper around the Go Standard Library sync.Pool primitive, which ensures ensures that we keep a static amount of values in memory (unaffected by GC), while allowing us the full benefits of the sync.Pool.

## Why

The Go author's are a lot smarter than me, so I don't think I could make a more performant sync.Pool (unlike other people). However, I do believe there are situations where you don't want your pool affected by GC. Somebody nerd-sniped me about this, so I built a little implementation here. We get the full capability of sync.Pool:

- Concurrency Safety
- Garbage Collection of the pool
- Pool item reuse
- Fast and standard

While getting the added bonus of being able to set a certain number of values as static (i.e. not GC-able).