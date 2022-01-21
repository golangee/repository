package repository

import (
	"context"
	"github.com/golangee/repository/iter"
	"io"
)

// A BlobRepository provides a bunch of methods to manage a set of blobs (binary large objects) identified by unique identifiers and
// is thread safe. An ID must not be a pointer type. Always use blobs for data you never want to hold in memory.
type BlobRepository[ID comparable] interface {
	Count(ctx context.Context) (int64, error)                 // Count enumerates all saved blobs at calling time. Due to concurrency, this is always only an indicator.
	Delete(ctx context.Context, id ID) error                  // Delete removes the given blob by id. It does not fail if no such ID exists.
	DeleteAll(ctx context.Context) error                      // DeleteAll clears the repository.
	Write(ctx context.Context, id ID) (io.WriteCloser, error) // Write allocates a new blob for the id and commits the written data on close. To stop, cancel the context.
	Read(ctx context.Context, id ID) (io.ReadCloser, error)   // Read opens the blob or returns a NotFoundError. Close to early release resources.
	FindAll(ctx context.Context) (iter.Iterator[ID], error)   // FindAll finds all blob ids.
}
