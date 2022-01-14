// Package iter provides a generic Iterator pattern, which is reduced to a minimum.
// This package is questionable and should be replaced until a stdlib contract approaches.
package iter

import (
	"errors"
	"io"
)

var (
	Done = errors.New("no more items")
)

// Iterator defines the minimal contract for an Iterator pattern.
// Implementations must deallocate or close any open resource (e.g. file or network) on the first returned error.
// If not all entries are consumed, underlying finalizers will clean any open file descriptors, however
// implementations may provide an optional io.Closer for immediate release.
type Iterator[T any] interface {
	// Next returns the next result. Returns Done if there are no more results. A return of Done is idempotent
	// and multiple calls have no effect.
	// See usage example.
	Next() (T, error)
}

// Walk applies a closure loop to the given iterator. The Iterator is closed if possible,
// if the closure returns an error early.
func Walk[T any](it Iterator[T], f func(item T) error) error {
	for {
		item, err := it.Next()
		if err == Done {
			break
		}

		if err != nil {
			return err
		}

		// check for early loop-cancellation by callback
		if err := f(item); err != nil {
			// try to release iterator immediately
			if c, ok := it.(io.Closer); ok {
				_ = c.Close() // suppress follow-up errors
			}

			return err // pass closure error
		}
	}

	return nil
}

// Collect loops the iterator and collects all items into a slice.
func Collect[T any](it Iterator[T]) ([]T, error) {
	var r []T
	for {
		item, err := it.Next()
		if err == Done {
			break
		}

		if err != nil {
			return r, err
		}

		r = append(r, item)
	}

	return r, nil
}

// Iter wraps a slice into an iterator.
func Iter[T any](slice []T) Iterator[T] {
	return &sliceIter[T]{buf: slice}
}

type sliceIter[T any] struct {
	pos int
	buf []T
}

func (s *sliceIter[T]) Next() (T, error) {
	var item T
	if s.pos >= len(s.buf) {
		return item, Done
	}

	item = s.buf[s.pos]
	s.pos++
	return item, nil
}
