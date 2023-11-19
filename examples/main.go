package main

import (
	"fmt"

	staticsyncpool "github.com/probably-not/static-sync-pool"
)

type ExampleStruct struct {
	A string
	B int64
	C float64
}

func main() {
	pool := staticsyncpool.New[ExampleStruct](
		func() *ExampleStruct {
			return &ExampleStruct{}
		},
		func(es *ExampleStruct) {
			es.A = ""
			es.B = 0
			es.C = 0.0
		},
		staticsyncpool.WithLazy(false),
		staticsyncpool.WithStaticSize(100),
	)

	s := pool.Get()

	fmt.Println("Received", s, "from pool")
}
