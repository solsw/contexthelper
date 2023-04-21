package contexthelper

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ContextAndContext returns [context.Context] combining two contexts with 'and' semantics.
func ContextAndContext(ctx1, ctx2 context.Context) context.Context {
	return NewAndContext(ctx1, ctx2)
}

// AndContext combines two contexts with 'and' semantics (see Done method).
// AndContext implements the [context.Context] interface.
type AndContext struct {
	Ctx1, Ctx2   context.Context
	onceDeadline sync.Once
	deadline     time.Time
	okDeadline   bool
	onceDone     sync.Once
	done         chan struct{}
	onceErr      sync.Once
	err          error
}

// check that AndContext implements the [context.Context] interface
var _ context.Context = &AndContext{}

// NewAndContext returns a new [AndContext].
func NewAndContext(ctx1, ctx2 context.Context) *AndContext {
	return &AndContext{Ctx1: ctx1, Ctx2: ctx2}
}

// Deadline implements the [context.Context.Deadline] method.
//
// If both deadlines are set, the latest one is returned.
func (c *AndContext) Deadline() (time.Time, bool) {
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
		if dl1.After(dl2) {
			c.deadline, c.okDeadline = dl1, true
			return
		}
		c.deadline, c.okDeadline = dl2, true
	})
	return c.deadline, c.okDeadline
}

func andDone(done1, done2 <-chan struct{}, done chan<- struct{}) {
	// done1 and done2 are not both nil here
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
//
// The returned channel is closed when both contexts' Done channels are closed.
func (c *AndContext) Done() <-chan struct{} {
	c.onceDone.Do(func() {
		if c.Ctx1.Done() == nil && c.Ctx2.Done() == nil {
			return
		}
		c.done = make(chan struct{})
		go andDone(c.Ctx1.Done(), c.Ctx2.Done(), c.done)
	})
	return c.done
}

// Err implements the [context.Context.Err] method.
//
// If [Done] is not yet closed, nil is returned.
// Otherwise an error that wraps non-nil contexts' Errs is returned.
func (c *AndContext) Err() error {
	select {
	case <-c.Done():
		c.onceErr.Do(func() { c.err = errors.Join(c.Ctx1.Err(), c.Ctx2.Err()) })
		return c.err
	default:
		return nil
	}
}

// Value implements the [context.Context.Value] method.
//
// If at least one Value method of combined contexts returns nil, nil is returned.
// Otherwise [TwoValues] struct containing values from both combined contexts is returned.
func (c *AndContext) Value(key any) any {
	v1 := c.Ctx1.Value(key)
	v2 := c.Ctx2.Value(key)
	if v1 == nil || v2 == nil {
		return nil
	}
	return TwoValues{v1, v2}
}
