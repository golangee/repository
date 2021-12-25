package repository

// TODO this is questionable and perhaps we should not do that until a stdlib approaches

const EOD constError = "end of data"

type constError string

func (e constError) Error() string {
	return string(e)
}

type Iterable[T any] interface {
	Iterator()
}

type Iterator[T any] interface {
	Next() (T, error)
	HasNext() bool
	Close() error // not so nice in go
	Error() error
}
