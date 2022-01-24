package fs

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golangee/repository/iter"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"sync"
)

var InvalidFilename = errors.New("invalid file name")
var byteType = reflect.TypeOf(byte(0))

const (
	MAX_PATH = 4096 - 1 // default linux limit excluding nul terminator
	NAME_MAX = 255
)

type Name interface {
	~string
}

type openFileMutex struct {
	mutex sync.RWMutex
	count int
}

// BlobRepository provides a simple fs based (not yet standardized) repository.
// This repository works on POSIX systems and just relies on the filesystem for concurrent and atomic
// read/write semantics which probably do not work on Windows.
// Behavior is undefined, if a directory is shared between multiple repository instances.
// The binary bytes of the ID is taken, hex encoded and used as the file name.
// This works best with plain integers or byte arrays (like UUID).
// Storing plain strings works, but is inefficient.
// Actually, the file is saved in a one-level fanout structure using the first byte of the sha256 hash
// of the encoded ID, to support repository sizes with a million objects (boils down to 4000 files per fanout dir):
//   hex(sha256(binary(id)))[0])/hex(binary(id))".bin"
// This implementation is mostly useful for prototyping and testing and shall not replace any serious SQL or NOSQL
// database. However, even though it may be slow, at least on POSIX it is considered to provide ACID properties.
type BlobRepository[ID Name] struct {
	fs   fs.FS
	pool *rcMutexes[ID]
}

func NewBlobRepository[ID Name](fsys fs.FS) (*BlobRepository[ID], error) {
	for prefix := byte(0); prefix < 0xff; prefix++ {
		if err := MkdirAll(fsys, hex.EncodeToString([]byte{prefix})); err != nil {
			return nil, fmt.Errorf("cannot initialize fanout: %w", err)
		}
	}

	return &BlobRepository[ID]{fs: fsys, pool: newRcMutexes[ID]()}, nil
}

func (r *BlobRepository[ID]) assertEmptyMutexes() {
	if len(r.pool.pool) != 0 {
		panic(fmt.Sprintf("expected empty pool but got %v entries", len(r.pool.pool)))
	}
}

func (r *BlobRepository[ID]) Count(ctx context.Context) (int64, error) {
	ids, err := r.FindAll(ctx)
	if err != nil {
		return 0, err
	}

	count := int64(0)
	err = iter.Walk(ids, func(item ID) error {
		count++
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *BlobRepository[ID]) Delete(ctx context.Context, id ID) error {
	if !ValidName(id) {
		return InvalidFilename
	}

	m := r.pool.get(id)
	m.inc()
	defer m.dec()

	m.Lock()
	defer m.Unlock()

	return Remove(r.fs, string(id))
}

func (r *BlobRepository[ID]) DeleteAll(ctx context.Context) error {
	ids, err := r.FindAll(ctx)
	if err != nil {
		return err
	}

	return iter.Walk(ids, func(item ID) error {
		return r.Delete(ctx, item)
	})
}

func (r *BlobRepository[ID]) Write(ctx context.Context, id ID) (io.WriteCloser, error) {
	if !ValidName(id) {
		return nil, InvalidFilename
	}

	return writeFile(r.fs, string(id), r.pool.get(id))

}

func (r *BlobRepository[ID]) Read(ctx context.Context, id ID) (io.ReadCloser, error) {
	if !ValidName(id) {
		return nil, InvalidFilename
	}

	return readFile(r.fs, string(id), r.pool.get(id))
}

// FindAll returns all blob identifiers. The current implementation buffers first the entire list of ids before
// the iterator becomes available. Files with a leading . are ignored.
func (r *BlobRepository[ID]) FindAll(ctx context.Context) (iter.Iterator[ID], error) {
	return r.FindByPrefix(ctx, ".")
}

// FindByPrefix is a special functions for this filesystem based implementation and allows to return
// a folder based prefix. To list the root, use '.' otherwise any ValidName denoting a directory is allowed.
// All contained files are returned recursively. Files with a leading . are ignored.
func (r *BlobRepository[ID]) FindByPrefix(ctx context.Context, prefix string) (iter.Iterator[ID], error) {
	if prefix != "." {
		if !ValidName(prefix) {
			return nil, InvalidFilename
		}
	}

	var res []ID
	err := fs.WalkDir(r.fs, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		if !d.IsDir() && !strings.HasSuffix(d.Name(), ".") {
			res = append(res, ID(path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return iter.Iter(res), nil
}

// ValidName returns false, if name does not apply to our rules of a safe name:
//  * 255 bytes per segment (NAME_MAX)
//  * no multibyte, no unicode, only a-z | 0-9 | . | _ | - to avoid normalization case sensitivity issues
//  * at most 4096 (MAX_PATH - 1) // including nul
//  * prefix / directory separator is /
// Our safe name rules represent more a or less the lowest subset of file names, which most common operating systems
// and storage apis support. This rules should be safe for windows, macos, linux and S3. For sure, Windows has
// some funny reserved names like LPT and COM etc. which are not checked.
func ValidName[T Name](name T) bool {
	if len(name) == 0 || len(name) > MAX_PATH {
		return false
	}

	for {
		i := 0
		for i < len(name) && name[i] != '/' {
			i++
		}

		elem := name[:i]
		if elem == "" || elem == "." || elem == ".." {
			return false
		}

		if len(elem) > NAME_MAX {
			return false
		}

		for c := 0; c < len(elem); c++ {
			v := elem[c]
			if !((v >= 'a' && v <= 'z') || (v >= '0' && v <= '9') || v == '.' || v == '_' || v == '-') {
				return false
			}
		}

		if i == len(name) {
			return true
		}

		name = name[i+1:]
	}
}
