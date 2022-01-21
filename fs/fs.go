package fs

import (
	"errors"
	"io"
	"io/fs"
)

var MkDirNotSupported = errors.New("fs does not support mkdir")
var RemoveNotSupported = errors.New("fs does not support remove")
var FileOpenNotSupported = errors.New("fs does not support OpenFile")
var WriteableFileNotSupported = errors.New("fs file does not write")
var RenameFileNotSupported = errors.New("fs does not support rename")

type RenameFileFS interface {
	fs.FS

	// Rename tries to perform an atomic rename if possible.
	Rename(oldpath, newpath string) error
}

type RemoveFileFS interface {
	fs.FS

	// Remove unlinks or deletes the given file or directory.
	Remove(name string) error
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

type OpenFileFS interface {
	fs.FS

	// OpenFile is the Posix-style fopen thing.
	OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
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

// Remove tries to remove the named file.
func Remove(fsys fs.FS, name string) error {
	if fsys, ok := fsys.(RemoveFileFS); ok {
		return fsys.Remove(name)
	}

	return RemoveNotSupported
}

// OpenFile tries open a file using posix style.
func OpenFile(fsys fs.FS, name string, flag int, perm fs.FileMode) (fs.File, error) {
	if fsys, ok := fsys.(OpenFileFS); ok {
		return fsys.OpenFile(name, flag, perm)
	}

	return nil, FileOpenNotSupported
}

func Rename(fsys fs.FS, oldpath, newpath string) error {
	if fsys, ok := fsys.(RenameFileFS); ok {
		return fsys.Rename(oldpath, newpath)
	}

	return RenameFileNotSupported
}
