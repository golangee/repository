package reflect

import "testing"

type MyType struct {
	Id   int
	Blub string
}

// go test -v -bench=. -benchmem ./...
func BenchmarkAllocValueType(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var t MyType
		Alloc[MyType](&t)
	}
}

func BenchmarkAllocPtrType(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var t *MyType
		Alloc[*MyType](&t)
	}
}

func BenchmarkConstructorValueType(b *testing.B) {
	cons, _ := Constructor[MyType]() // this is zero alloc, just stack copying
	for n := 0; n < b.N; n++ {
		var t MyType = cons()
		t.Id = n
		_ = t
	}
}

func BenchmarkConstructorPtrType(b *testing.B) {
	cons, _ := Constructor[*MyType]()
	for n := 0; n < b.N; n++ {
		var t *MyType = cons()
		t.Id = n
		_ = t
	}
}
