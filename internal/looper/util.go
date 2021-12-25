package looper

func FindAll[T any, ID comparable](r interface {
	FindAll(consumer func(ID, T) error) error
}) ([]T, error) {
	var res []T
	r.FindAll(func(id ID, t T) error {
		res = append(res, t)
		return nil
	})

	return res, nil
}
