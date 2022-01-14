package fs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golangee/repository/iter"
	"io"
	"io/fs"
	"io/ioutil"
	"path"
	"reflect"
	"strconv"
	"strings"
)

const blobFileExt = ".bin"

var InvalidFilename = errors.New("invalid file name")
var byteType = reflect.TypeOf(byte(0))

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
type BlobRepository[ID any] struct {
	fs fs.FS
}

func NewBlobRepository[ID any](fsys fs.FS) (*BlobRepository[ID], error) {
	for prefix := byte(0); prefix < 0xff; prefix++ {
		if err := MkdirAll(fsys, hex.EncodeToString([]byte{prefix})); err != nil {
			return nil, fmt.Errorf("cannot initialize fanout: %w", err)
		}
	}

	return &BlobRepository[ID]{fs: fsys}, nil
}

func (r *BlobRepository[ID]) Count(ctx context.Context) (int64, error) {
	ids, err := r.Blobs(ctx)
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
	//TODO implement me
	panic("implement me")
}

func (r *BlobRepository[ID]) DeleteAll(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (r *BlobRepository[ID]) Write(ctx context.Context, id ID) (io.WriteCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (r *BlobRepository[ID]) Read(ctx context.Context, id ID) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

// Blobs returns all blob identifiers. The current implementation buffers first the entire list of ids before
// the iterator becomes available.
func (r *BlobRepository[ID]) Blobs(ctx context.Context) (iter.Iterator[ID], error) {
	var res []ID
	err := fs.WalkDir(r.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), blobFileExt) {
			id, err := decode[ID](d.Name())
			if err != nil {
				return InvalidFilename
			}

			res = append(res, id)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return iter.Iter(res), nil
}

// decode is the reverse of encode.
func decode[ID any](filename string) (ID, error) {
	fname := path.Base(filename)
	ext := path.Ext(fname)
	if ext != "" {
		fname = fname[:len(fname)-len(ext)]
	}

	var id ID // only values types supported
	dec := base32.NewDecoder(base32.HexEncoding.WithPadding(base32.NoPadding), bytes.NewReader([]byte(fname)))
	buf, err := ioutil.ReadAll(dec)
	if err != nil {
		return id, fmt.Errorf("invalid base32 encoding in filename: %s (%s): %w", fname, filename, err)
	}

	err = json.Unmarshal(buf, &id)
	if err != nil {
		return id, fmt.Errorf("cannot unmarshal json encoded filename: %s: %w", fname, err)
	}

	return id, nil
}

// encode returns the filename for the id:
//  hex(sha256(binary(id)))[0])/hex(binary(id))".bin"
func encode[ID any](id ID) (string, error) {
	// currently, we are screwed, see https://github.com/golang/go/issues/45380
	var buf []byte
	t := reflect.TypeOf(id)
	v := reflect.ValueOf(id)
	safeString := false
	switch t.Kind() {
	case reflect.Array:
		if t.Elem() == byteType {
			fmt.Println("=> got a byte array", t.Elem(), t)
			buf = make([]byte, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				buf = append(buf, byte(v.Index(i).Uint()))
			}
		}
	case reflect.Slice:
		if t.Elem() == byteType {
			fmt.Println("=> got a byte slice", t.Elem(), t)
			buf = v.Bytes()
		}

	case reflect.String:
		// actually we could just allow that, but that seems to cause a lot of headache:
		// - we must check for directory traversal or injection attacks
		// - file systems causing collisions due to case insensitivity
		// - file systems causing missing identifiers due to hidden unicode normalization (windows, macos)
		// - file systems not supporting all byte sequences (windows, macos)
		buf = []byte(any(id).(string))
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		buf = strconv.AppendInt(buf, v.Int(), 16)
		safeString = true

	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		buf = strconv.AppendUint(buf, v.Uint(), 16)
		safeString = true
	}

	if buf == nil {
		panic("unsupported id type for encoding: " + t.String())
	}

	fmt.Println(id, "=>", buf)

	sum := sha256.Sum256(buf)
	fanout := hex.EncodeToString(sum[:1])
	var fname string
	if safeString {
		fname = string(buf)
	} else {
		fname = hex.EncodeToString(buf)
	}
	return path.Join(fanout, fname+blobFileExt), nil
}
