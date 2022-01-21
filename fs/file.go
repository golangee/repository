package fs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"sync"
	"time"
)

type rcMutexes[K comparable] struct {
	pool map[K]*rcMutex
	lock sync.Mutex
}

func newRcMutexes[K comparable]() *rcMutexes[K] {
	return &rcMutexes[K]{pool: map[K]*rcMutex{}}
}

func (m *rcMutexes[K]) get(k K) *rcMutex {
	m.lock.Lock()
	defer m.lock.Unlock()

	mutex := m.pool[k]
	if mutex == nil {
		mutex = &rcMutex{
			destroy: func(self *rcMutex) {
				m.lock.Lock()
				defer m.lock.Unlock()

				delete(m.pool, k)
			},
		}
		m.pool[k] = mutex
	}

	return mutex
}

type rcMutex struct {
	sync.RWMutex
	rcLock  sync.Mutex
	rc      int
	destroy func(self *rcMutex)
}

func (r *rcMutex) inc() {
	r.rcLock.Lock()
	defer r.rcLock.Unlock()
	r.rc++
}

func (r *rcMutex) dec() {
	r.rcLock.Lock()
	defer r.rcLock.Unlock()
	r.rc--

	if r.rc < 0 {
		panic("invalid reference count for mutex")
	}

	if r.rc == 0 {
		r.destroy(r)
	}
}

// fileReadCloser only provides concurrent read access.
type fileReadCloser struct {
	mutex *rcMutex
	file  fs.File
}

func readFile(fsys fs.FS, name string, mutex *rcMutex) (*fileReadCloser, error) {
	mutex.inc()
	mutex.RLock() // lock before, to avoid races
	file, err := OpenFile(fsys, name, os.O_RDONLY, 0)
	if err != nil {
		mutex.RUnlock() // unlock, e.g. file does not exist
		mutex.dec()
		return nil, err
	}

	return &fileReadCloser{
		mutex: mutex,
		file:  file,
	}, nil
}

func (f *fileReadCloser) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f *fileReadCloser) Close() error {
	defer f.mutex.dec()
	defer f.mutex.RUnlock()
	if closer, ok := f.file.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

// fileWriteCloser writes into a temporary file and locks the file writeable only when committing, forcing
// any other read locks to close before. This ensures most portable cross-platform behavior for atomic renames,
// especially on systems without posix unlink semantic like windows.
type fileWriteCloser struct {
	mutex   *rcMutex
	fsys    fs.FS
	dstName string
	tmpName string
	tmpFile WriteableFile
}

func writeFile(fsys fs.FS, name string, mutex *rcMutex) (*fileWriteCloser, error) {
	mutex.inc() // ensure mutex live time
	tmpName := "." + name + "." + strconv.FormatInt(time.Now().UnixMicro(), 10) + ".tmp"
	file, err := OpenFile(fsys, tmpName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		mutex.dec()
		return nil, err
	}

	if w, ok := file.(WriteableFile); ok {
		return &fileWriteCloser{
			mutex:   mutex,
			tmpName: tmpName,
			tmpFile: w,
			dstName: name,
			fsys:    fsys,
		}, nil
	}

	return nil, WriteableFileNotSupported
}

func (f *fileWriteCloser) Write(p []byte) (n int, err error) {
	return f.Write(p)
}

func (f *fileWriteCloser) Close() error {
	defer f.mutex.dec() //free mutex

	if syncer, ok := f.tmpFile.(SyncableFile); ok {
		if err := syncer.Sync(); err != nil {
			return fmt.Errorf("fsync failed on temporary file: %w", err)
		}
	}

	if err := f.tmpFile.Close(); err != nil {
		return fmt.Errorf("cannot close temporary file: %w", err)
	}

	f.mutex.Lock() // acquire the write-lock, waiting that all readers are closed on shared mutex
	defer f.mutex.Unlock()

	if err := Rename(f.fsys, f.tmpName, f.dstName); err != nil {
		return fmt.Errorf("cannot rename file %s -> %s: %w", f.tmpName, f.dstName, err)
	}

	return nil
}
