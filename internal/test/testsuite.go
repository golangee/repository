package test

import (
	"errors"
	"fmt"
	"github.com/golangee/repository/internal/looper"
	"io"
	"reflect"
	"testing"
)

// CrudTestRepository avoids defining a circular dependency between this and the testing packages.
type CrudTestRepository[T any, ID comparable] interface {
	Count() (int64, error)
	DeleteByID(id ID) error
	DeleteAll() error
	Save(id ID, entity T) error
	SaveAll(producer func() (ID, T, error)) error
	FindByID(id ID) (T, error)
	FindAll(consumer func(ID, T) error) error
}

type TestTableEntry[T any, ID comparable] struct {
	ID     ID
	Entity T
}

type A string
type B struct {
	ID        string
	Firstname string
	Age       int
	Address   []struct {
		Street string
		Zip    string
	}
}

func CreateTestSet1() []TestTableEntry[A, string] {
	return []TestTableEntry[A, string]{
		{"1", ""},
		{"", "a"},
		{"2", "abc"},
		{"/\\:", `:'"!`},
		{"a b c", "hello world"},
	}
}

func Test[T any, ID comparable](t *testing.T, table []TestTableEntry[T, ID], repo CrudTestRepository[T, ID]) {
	t.Helper()
	assert(expect(repo.Count()), 0)
	must(repo.DeleteAll())
	assert(expect(repo.Count()), 0)

	// insert one after another
	for i, e := range table {
		must(repo.Save(e.ID, e.Entity))
		assert(expect(repo.Count()), int64(i)+1)
		clone := expect(repo.FindByID(e.ID))
		assert(clone, e.Entity)

		all := expect(looper.FindAll[T, ID](repo))
		assert(len(all), i+1)
	}

	// overwrite
	for _, e := range table {
		must(repo.Save(e.ID, e.Entity))
		assert(expect(repo.Count()), int64(len(table)))
		clone := expect(repo.FindByID(e.ID))
		assert(clone, e.Entity)

		all := expect(looper.FindAll[T, ID](repo))
		assert(len(all), len(table))
	}

	// delete one after another
	for i, e := range table {
		must(repo.DeleteByID(e.ID))

		_, err := repo.FindByID(e.ID)

		var notFound interface{ NotFound() bool }
		if !errors.As(err, &notFound) {
			t.Fatal("expected not found")
		}

		assert(expect(repo.Count()), int64(len(table)-i-1))
		assert(len(expect(looper.FindAll[T, ID](repo))), len(table)-i-1)
	}

	// save all
	idx := 0
	repo.SaveAll(func() (ID, T, error) {
		defer func() { idx++ }()
		if idx >= len(table) {
			var zID ID
			var zT T
			return zID, zT, io.EOF
		}

		return table[idx].ID, table[idx].Entity, nil
	})

	// re-read again
	for _, e := range table {
		clone := expect(repo.FindByID(e.ID))
		assert(clone, e.Entity)
	}

	// delete all
	must(repo.DeleteAll())
	assert(expect(repo.Count()), 0)

	t.Log("test suite pass:", reflect.TypeOf(repo).String())
}

func must(err error) {
	expect[any](nil, err)
}

func expect[T any](t T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("error not expected: %v", err))
	}

	return t
}

// even though generics do not add much, this makes Go more save and clever here:
//  * generics enforce the same type, so the inferring performs the correct type conversion (0 -> int64 instead of int)
func assert[T any](actual, expected T) {
	if !reflect.DeepEqual(actual, expected) {
		panic(fmt.Sprintf("expected %v but got %v", expected, actual))
	}
}
