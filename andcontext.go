package contexthelper

import (
	"context"
	"sync"
	"time"

	"github.com/solsw/generichelper"
)

// AndContext combines two contexts with AND semantics (see [AndContext.Done] method).
// AndContext implements the [context.Context] interface.
type AndContext struct {
	ctx1, ctx2   context.Context
	onceDeadline sync.Once
	deadline     time.Time
	okDeadline   bool
	onceDone     sync.Once
	done         <-chan struct{}
	onceErr      sync.Once
	err          error
}

// check that &AndContext implements the [context.Context] interface
var _ context.Context = &AndContext{}

// NewAndContext returns a new [AndContext].
func NewAndContext(ctx1, ctx2 context.Context) *AndContext {
	return &AndContext{ctx1: ctx1, ctx2: ctx2}
}

// Deadline implements the [context.Context.Deadline] method.
// If both deadlines are set, the latest one is returned.
// A sole deadline applies only when the other context has a nil Done channel.
func (c *AndContext) Deadline() (time.Time, bool) {
	c.onceDeadline.Do(func() {
		dl1, ok1 := c.ctx1.Deadline()
		dl2, ok2 := c.ctx2.Deadline()
		switch {
		case ok1 && ok2:
			c.deadline, c.okDeadline = dl2, true
			if dl1.After(dl2) {
				c.deadline = dl1
			}
		case ok1 && c.ctx2.Done() == nil:
			c.deadline, c.okDeadline = dl1, true
		case ok2 && c.ctx1.Done() == nil:
			c.deadline, c.okDeadline = dl2, true
		}
	})
	return c.deadline, c.okDeadline
}

func andDone(done1, done2 <-chan struct{}, done chan<- struct{}) {
	for done1 != nil || done2 != nil {
		select {
		case _, ok1 := <-done1:
			if !ok1 {
				done1 = nil
			}
		case _, ok2 := <-done2:
			if !ok2 {
				done2 = nil
			}
		}
	}
	close(done)
}

// Done implements the [context.Context.Done] method.
// The return channel is closed when all non-nil Done channels of the combined contexts are closed.
// A context with a nil Done channel (e.g. [context.Background]) is treated as already done.
func (c *AndContext) Done() <-chan struct{} {
	c.onceDone.Do(func() {
		c.done = combineDone(c.ctx1.Done(), c.ctx2.Done(), andDone)
	})
	return c.done
}

// Err implements the [context.Context.Err] method.
// If [AndContext.Done] is not yet closed, nil is returned.
// A sole non-nil [Err] is returned directly. If both contexts have errors,
// the returned error [wraps] both.
//
// [Err]: https://pkg.go.dev/context#Context.Err
// [wraps]: https://pkg.go.dev/errors#Join
func (c *AndContext) Err() error {
	select {
	case <-c.Done():
		c.onceErr.Do(func() { c.err = combineErrors(c.ctx1.Err(), c.ctx2.Err()) })
		return c.err
	default:
		return nil
	}
}

// Value implements the [context.Context.Value] method.
// If neither combined context holds 'key', nil is returned.
// Otherwise, a [generichelper.Tuple2][any, any] is returned whose items are the values
// from each combined context; an item is nil if the corresponding context does not hold 'key'.
func (c *AndContext) Value(key any) any {
	v1 := c.ctx1.Value(key)
	v2 := c.ctx2.Value(key)
	if v1 == nil && v2 == nil {
		return nil
	}
	return generichelper.Tuple2[any, any]{Item1: v1, Item2: v2}
}
