package repository

import "fmt"

type Entity[ID comparable] interface {
	GetID() ID
}

// A CrudRepository provides a bunch of methods to manage a set of entities identified by unique identifiers and
// is thread safe. The ownership of any returned Entity instance is handed over to the callee and never reused internally.
// This is quite inefficient but always thread safe.
type CrudRepository[T Entity[ID], ID comparable] interface {
	Count() (int64, error)         // Count enumerates all saved entities at calling time.
	DeleteByID(id ID) error        // DeleteByID remove the given entity. It does not fail if no such ID exists.
	DeleteAll() error              // DeleteAll clears the repository.
	Save(entity T) error           // Save overwrites the entity identified by its ID. It does not fail whether entity already exists.
	FindByID(id ID) (T, error)     // FindByID returns either T or EntityNotFoundError.
	FindAll(f func(T) error) error // FindAll invokes the callback for each entity.
}

type EntityNotFoundError struct {
	ID any
}

func (e EntityNotFoundError) GetID() any {
	return e.ID
}

func (e EntityNotFoundError) EntityNotFoundError() bool {
	return true
}

func (e EntityNotFoundError) Error() string {
	return fmt.Sprintf("entity not found: %v", e.ID)
}
