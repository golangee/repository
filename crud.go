package repository

import (
	"fmt"
	"github.com/golangee/repository/internal/looper"
)

// A CrudRepository provides a bunch of methods to manage a set of entities identified by unique identifiers and
// is thread safe. The ownership of any returned Entity instance is handed over to the callee and never reused internally.
// This is quite inefficient but always thread safe.
type CrudRepository[T any, ID comparable] interface {
	Count() (int64, error)                        // Count enumerates all saved entities at calling time.
	DeleteByID(id ID) error                       // DeleteByID remove the given entity. It does not fail if no such ID exists.
	DeleteAll() error                             // DeleteAll clears the repository.
	Save(id ID, entity T) error                   // Save overwrites the entity identified by its ID. It does not fail whether entity already exists.
	SaveAll(producer func() (ID, T, error)) error // SaveAll stores all entities until the first error occurs. Returning an io.EOF will finish processing.
	FindByID(id ID) (T, error)                    // FindByID returns either T or EntityNotFoundError.
	FindAll(consumer func(ID, T) error) error     // FindAll invokes the callback for each entity. The order is unspecified.
}

func FindAll[T any, ID comparable](r interface {
	FindAll(consumer func(ID, T) error) error
}) ([]T, error) {
	return looper.FindAll(r)
}

type EntityNotFoundError struct {
	ID any
}

func (e EntityNotFoundError) GetID() any {
	return e.ID
}

func (e EntityNotFoundError) NotFound() bool {
	return true
}

func (e EntityNotFoundError) Error() string {
	return fmt.Sprintf("entity not found: %v", e.ID)
}
