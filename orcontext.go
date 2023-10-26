package contexthelper

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/solsw/generichelper"
)

// ContextOrContext returns [context.Context] combining two contexts with OR semantics.
func ContextOrContext(ctx1, ctx2 context.Context) context.Context {
	return NewOrContext(ctx1, ctx2)
}

// OrContext combines two contexts with OR semantics (see [OrContext.Done] method).
// OrContext implements the [context.Context] interface.
type OrContext struct {
	Ctx1, Ctx2   context.Context
	onceDeadline sync.Once
	deadline     time.Time
	okDeadline   bool
	onceDone     sync.Once
	done         chan struct{}
	onceErr      sync.Once
	err          error
}

// check that &OrContext implements the [context.Context] interface
var _ context.Context = &OrContext{}

// NewOrContext returns a new [OrContext].
func NewOrContext(ctx1, ctx2 context.Context) *OrContext {
	return &OrContext{Ctx1: ctx1, Ctx2: ctx2}
}

// Deadline implements the [context.Context.Deadline] method.
// If both deadlines are set, the earliest one is returned.
func (c *OrContext) Deadline() (time.Time, bool) {
	c.onceDeadline.Do(func() {
		dl1, ok1 := c.Ctx1.Deadline()
		dl2, ok2 := c.Ctx2.Deadline()
		if !ok1 {
			c.deadline, c.okDeadline = dl2, ok2
			return
		}
		if !ok2 {
			c.deadline, c.okDeadline = dl1, ok1
			return
		}
		if dl1.Before(dl2) {
			c.deadline, c.okDeadline = dl1, true
			return
		}
		c.deadline, c.okDeadline = dl2, true
	})
	return c.deadline, c.okDeadline
}

func orDone(done1, done2 <-chan struct{}, done chan<- struct{}) {
	// done1 and done2 are not both nil here
	select {
	case <-done1:
	case <-done2:
	}
	close(done)
}

// Done implements the [context.Context.Done] method.
// The return channel is closed when either one of contexts' Done channels is closed.
func (c *OrContext) Done() <-chan struct{} {
	c.onceDone.Do(func() {
		if c.Ctx1.Done() == nil && c.Ctx2.Done() == nil {
			return
		}
		c.done = make(chan struct{})
		go orDone(c.Ctx1.Done(), c.Ctx2.Done(), c.done)
	})
	return c.done
}

// Err implements the [context.Context.Err] method.
// If [Done] is not yet closed, nil is returned.
// Otherwise, an error that wraps non-nil contexts' Errs is returned.
func (c *OrContext) Err() error {
	select {
	case <-c.Done():
		c.onceErr.Do(func() { c.err = errors.Join(c.Ctx1.Err(), c.Ctx2.Err()) })
		return c.err
	default:
		return nil
	}
}

// Value implements the [context.Context.Value] method.
// If Value methods of both combined contexts return nil, nil is returned.
// Otherwise, [generichelper.Tuple2] struct containing values from both combined contexts is returned.
func (c *OrContext) Value(key any) any {
	v1 := c.Ctx1.Value(key)
	v2 := c.Ctx2.Value(key)
	if v1 == nil && v2 == nil {
		return nil
	}
	return generichelper.Tuple2[any, any]{Item1: v1, Item2: v2}
}
