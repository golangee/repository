package mem

import (
	"encoding/json"
	"github.com/golangee/repository"
	"reflect"
	"sync"
)

type Repository[T repository.Entity[ID], ID comparable] struct {
	mutex       sync.RWMutex
	store       map[ID][]byte
	pointerType bool
}

func NewRepository[T repository.Entity[ID], ID comparable]() *Repository[T, ID] {
	var zeroT T

	return &Repository[T, ID]{
		store:       map[ID][]byte{},
		pointerType: reflect.ValueOf(zeroT).Kind() == reflect.Ptr,
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

	var zeroT T
	buf, ok := r.store[id]
	if !ok {
		return zeroT, repository.EntityNotFoundError{ID: id}
	}

	if r.pointerType {
		// currently, there seems to be no generic way of creating a pointer instance
		x := reflect.New(reflect.TypeOf(zeroT).Elem())
		zeroT = x.Interface().(T)
		if err := json.Unmarshal(buf, zeroT); err != nil {
			return zeroT, err
		}
	} else {
		if err := json.Unmarshal(buf, &zeroT); err != nil {
			return zeroT, err
		}
	}

	return zeroT, nil
}

// FindAll invokes the callback for each entry and keeps the ownership.
// Calling any other instance method from the callback will cause a deadlock.
func (r *Repository[T, ID]) FindAll(f func(entity T) error) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, entity := range r.store {
		//if err := f(entity); err != nil {
		//	return err
		//}
		_ = entity
	}

	return nil
}
