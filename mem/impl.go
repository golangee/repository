package mem

import (
	"encoding/json"
	"github.com/golangee/repository"
	"github.com/golangee/repository/internal/reflect"

	"sync"
)

// Repository is a generic CrudRepository using json marshalling to deep clone the entities.
// Even though this is very demanding for an in-memory store, it guarantees data consistency
// and no data races when modifying the entities.
type Repository[T repository.Entity[ID], ID comparable] struct {
	mutex     sync.RWMutex
	store     map[ID][]byte
	factory   func() T
	isPtrType bool
}

func NewRepository[T repository.Entity[ID], ID comparable]() *Repository[T, ID] {
	fac, ptr := reflect.Constructor[T]()

	return &Repository[T, ID]{
		store:     map[ID][]byte{},
		factory:   fac,
		isPtrType: ptr,
	}
}

func (r *Repository[T, ID]) Count() (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return int64(len(r.store)), nil
}

func (r *Repository[T, ID]) DeleteByID(id ID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.store, id)
	return nil
}

func (r *Repository[T, ID]) DeleteAll() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// intentionally releasing old map to also free potential large backing slices
	r.store = map[ID][]byte{}

	return nil
}

func (r *Repository[T, ID]) Save(entity T) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	buf, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	r.store[entity.GetID()] = buf
	return nil
}

func (r *Repository[T, ID]) FindByID(id ID) (T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var entity T
	buf, ok := r.store[id]
	if !ok {
		return entity, repository.EntityNotFoundError{ID: id}
	}

	return r.unmarshal(buf)
}

// FindAll invokes the callback for each entry and transfers the ownership.
// Calling any other instance method from the callback will cause a deadlock.
func (r *Repository[T, ID]) FindAll(f func(entity T) error) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, buf := range r.store {
		entity, err := r.unmarshal(buf)
		if err != nil {
			return err
		}

		if err := f(entity); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository[T, ID]) unmarshal(buf []byte) (T, error) {
	entity := r.factory()
	if r.isPtrType {
		if err := json.Unmarshal(buf, entity); err != nil {
			return entity, err
		}
	} else {
		if err := json.Unmarshal(buf, &entity); err != nil {
			return entity, err
		}
	}

	return entity, nil
}
