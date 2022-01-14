package iter_test

import (
	"fmt"
	"github.com/golangee/repository/iter"
)

func ExampleIterator_next() {
	var it iter.Iterator[string]
	for {
		item, err := it.Next()
		if err == iter.Done {
			break
		}

		if err != nil {
			// TODO: Handle error.
		}

		fmt.Println(item)
	}
}

func ExampleIterator_next2() {
	var it iter.Iterator[string]
	for item, err := it.Next(); err != iter.Done; {

		if err != nil {
			// TODO: Handle error.
		}

		fmt.Println(item)
	}
}
