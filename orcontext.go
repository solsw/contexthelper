package contexthelper

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ContextOrContext returns [context.Context] combining two contexts with 'or' semantics.
func ContextOrContext(ctx1, ctx2 context.Context) context.Context {
	return NewOrContext(ctx1, ctx2)
}

// OrContext combines two contexts with 'or' semantics (see Done method).
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

// check that OrContext implements the [context.Context] interface
var _ context.Context = &OrContext{}

// NewOrContext returns a new OrContext.
func NewOrContext(ctx1, ctx2 context.Context) *OrContext {
	return &OrContext{Ctx1: ctx1, Ctx2: ctx2}
}

// Deadline implements the [context.Context.Deadline] method.
//
// If both deadlines are set, the earliest one is returned.
func (cc *OrContext) Deadline() (time.Time, bool) {
	cc.onceDeadline.Do(func() {
		dl1, ok1 := cc.Ctx1.Deadline()
		dl2, ok2 := cc.Ctx2.Deadline()
		if !ok1 {
			cc.deadline, cc.okDeadline = dl2, ok2
			return
		}
		if !ok2 {
			cc.deadline, cc.okDeadline = dl1, ok1
			return
		}
		if dl1.Before(dl2) {
			cc.deadline, cc.okDeadline = dl1, true
			return
		}
		cc.deadline, cc.okDeadline = dl2, true
	})
	return cc.deadline, cc.okDeadline
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
//
// The returned channel is closed when either one of contexts' Done channels is closed.
func (cc *OrContext) Done() <-chan struct{} {
	cc.onceDone.Do(func() {
		if cc.Ctx1.Done() == nil && cc.Ctx2.Done() == nil {
			return
		}
		cc.done = make(chan struct{})
		go orDone(cc.Ctx1.Done(), cc.Ctx2.Done(), cc.done)
	})
	return cc.done
}

// Err implements the [context.Context.Err] method.
//
// If both contexts' Errs are nil, nil is returned.
// Otherwise an error that wraps non-nil contexts' Errs is returned.
func (cc *OrContext) Err() error {
	select {
	case <-cc.Done():
		cc.onceErr.Do(func() { cc.err = errors.Join(cc.Ctx1.Err(), cc.Ctx2.Err()) })
		return cc.err
	default:
		return nil
	}
}

// Value implements the [context.Context.Value] method.
//
// The method returns nil.
func (*OrContext) Value(key any) any {
	return nil
}
