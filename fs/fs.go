package fs

import (
	"errors"
	"io"
	"io/fs"
)

var WriteNotSupported = errors.New("fs does not support file writes")
var MkDirNotSupported = errors.New("fs does not support mkdir")

// WriteFileFS is the interface implemented by a file system
// that provides an optimized implementation of WriteFile.
type WriteFileFS interface {
	fs.FS

	// WriteFile writes the named file.
	WriteFile(name string, data []byte) error
}

// WriteFS is the interface implemented by a file system
// that provides an optimized implementation of Write.
type WriteFS interface {
	fs.FS

	// Write writes the named file.
	Write(name string, w func(w io.Writer) error) error
}

// MakeDirFileFS is the interface implemented by a file system
// that provides an optimized implementation of WriteFile.
type MakeDirFileFS interface {
	fs.FS

	// MkdirAll creates all folders, if required. If name denotes already directories, returns nil.
	MkdirAll(name string) error
}

type WriteableFile interface {
	fs.File
	io.Writer
}

type SyncableFile interface {
	WriteableFile
	Sync() error
}

// MkdirAll tries to create the given hierarchy of directories.
func MkdirAll(fsys fs.FS, name string) error {
	if fsys, ok := fsys.(MakeDirFileFS); ok {
		return fsys.MkdirAll(name)
	}

	return MkDirNotSupported
}

// WriteFile writes the named file into the file system fs replacing and truncating its content.
// A successful call returns a nil error.
//
// If fs implementsWriteFileFS, WriteFile calls fs.WriteFile.
// Otherwise, WriteFile calls fs.Open and uses Write, Sync (optionally) and Close
// on the returned file.
func WriteFile(fsys fs.FS, name string, data []byte) (err error) {
	if fsys, ok := fsys.(WriteFileFS); ok {
		return fsys.WriteFile(name, data)
	}

	file, err := fsys.Open(name)
	if err != nil {
		return err
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	wfile, ok := file.(WriteableFile)
	if !ok {
		return WriteNotSupported
	}

	if _, err := wfile.Write(data); err != nil {
		return err
	}

	if file, ok := file.(SyncableFile); ok {
		if err := file.Sync(); err != nil {
			return err
		}
	}

	return // see defer close
}
