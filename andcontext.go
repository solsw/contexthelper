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

// NewAndContext returns a new AndContext.
func NewAndContext(ctx1, ctx2 context.Context) *AndContext {
	return &AndContext{Ctx1: ctx1, Ctx2: ctx2}
}

// Deadline implements the [context.Context.Deadline] method.
//
// If both deadlines are set, the latest one is returned.
func (cc *AndContext) Deadline() (time.Time, bool) {
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
		if dl1.After(dl2) {
			cc.deadline, cc.okDeadline = dl1, true
			return
		}
		cc.deadline, cc.okDeadline = dl2, true
	})
	return cc.deadline, cc.okDeadline
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
func (cc *AndContext) Done() <-chan struct{} {
	cc.onceDone.Do(func() {
		if cc.Ctx1.Done() == nil && cc.Ctx2.Done() == nil {
			return
		}
		cc.done = make(chan struct{})
		go andDone(cc.Ctx1.Done(), cc.Ctx2.Done(), cc.done)
	})
	return cc.done
}

// Err implements the [context.Context.Err] method.
//
// If both contexts' Errs are nil, nil is returned.
// Otherwise an error that wraps non-nil contexts' Errs is returned.
func (cc *AndContext) Err() error {
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
func (*AndContext) Value(key any) any {
	return nil
}
