package contexthelper

import (
	"context"
)

// Value returns the value of type T associated with this context for 'key'
// and a bool indicating whether the value exists and is of T type.
func Value[T any](ctx context.Context, key any) (T, bool) {
	v, ok := ctx.Value(key).(T)
	return v, ok
}
