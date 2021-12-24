package reflect

import "reflect"

// Alloc checks if T is a pointer type and allocates an appropriate T for it.
// This code always causes expensive escaping behavior.
func Alloc[T any](p *T) (isPtr bool) {
	zero := reflect.ValueOf(p).Elem()
	isPtr = zero.Kind() == reflect.Ptr
	if !isPtr {
		return // nothing to do, pointing to value type
	}

	x := reflect.New(zero.Type().Elem())
	t := x.Interface().(T)
	*p = t

	return
}

// Constructor allows a zero allocation generic value type (stack only) creation and a
// "single" heap allocation for pointer types. Keep the constructor for efficiency.
func Constructor[T any]() (factory func() T, isPtrType bool) {
	var zeroT T
	isPtr := reflect.TypeOf(zeroT).Kind() == reflect.Ptr // allocates only once
	return func() T {
		if !isPtr {
			return zeroT
		}

		var localT T
		return reflect.New(reflect.TypeOf(localT).Elem()).Interface().(T)
	}, isPtr
}
