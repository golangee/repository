package fs

import (
	"encoding/hex"
	"fmt"
	"github.com/golangee/repository/internal/reflect"
	"io/fs"
)

// Repository is a generic CrudRepository using json marshalling to serialize into the filesystem.

type Repository[T any, ID comparable] struct {
	factory   func() T
	isPtrType bool
	fs        fs.FS
}

func NewRepository[T any, ID comparable](fs fs.FS) (*Repository[T, ID], error) {
	fac, ptr := reflect.Constructor[T]()

	for prefix := byte(0); prefix < 0xff; prefix++ {
		if err := MkdirAll(fs, hex.EncodeToString([]byte{prefix})); err != nil {
			return nil, fmt.Errorf("cannot initialize fanout: %w", err)
		}
	}

	return &Repository[T, ID]{
		factory:   fac,
		isPtrType: ptr,
		fs:        fs,
	}, nil
}
