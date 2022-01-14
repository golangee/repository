package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type dirFS struct {
	root string
}

// Dir opens a read and writeable filesystem which implements a WriteFile method with transactional semantics.
func Dir(path string) fs.FS {
	return dirFS{root: path}
}

func (l dirFS) Open(name string) (fs.File, error) {
	if name == "." {
		name = ""
	}

	// windows also accepts the slashes from fs.FS
	return os.Open(filepath.Join(l.root, name))
}

func (l dirFS) MkdirAll(name string) error {
	// windows also accepts the slashes from fs.FS
	return os.MkdirAll(filepath.Join(l.root, name), 0700) // just allow owner read/write/list and not the world
}

// WriteFile performs a transactional write, a fsync and an atomic posix rename.
// Fails on windows, if destination file is still open, posix allows that kind of concurrency.
func (l dirFS) WriteFile(name string, data []byte) (err error) {
	return l.Write(name, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// Write performs a transactional write, a fsync and an atomic posix rename.
// Fails on windows, if destination file is still open, posix allows that kind of concurrency.
func (l dirFS) Write(name string, w func(io.Writer) error) (err error) {
	// windows also accepts the slashes from fs.FS
	dst := filepath.Join(l.root, name)
	tmp := dst + "." + strconv.FormatInt(time.Now().UnixMicro(), 10) + ".tmp"

	// just allow owner read/write and not the world
	file, e := os.OpenFile(tmp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}

	defer func() {
		e := file.Close() // delayed write may fail (e.g. fuse, nfs, etc)
		if e != nil {
			if err == nil {
				err = e
			}

			return
		}

		// posix atomic replace
		if e := os.Rename(tmp, dst); e != nil && err == nil {
			err = e
		}
	}()

	if err := w(file); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return // see defer
}
