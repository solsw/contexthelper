package contexthelper

import (
	"context"
	"errors"
)

// Value returns the value of type T associated with 'ctx' for 'key'
// and a bool reporting whether the value exists and is of T type.
func Value[T any](ctx context.Context, key any) (T, bool) {
	v, ok := ctx.Value(key).(T)
	return v, ok
}

func combineDone(done1, done2 <-chan struct{}, wait func(<-chan struct{}, <-chan struct{}, chan<- struct{})) <-chan struct{} {
	if done1 == nil {
		return done2
	}
	if done2 == nil {
		return done1
	}
	done := make(chan struct{})
	go wait(done1, done2, done)
	return done
}

func combineErrors(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return errors.Join(err1, err2)
}
