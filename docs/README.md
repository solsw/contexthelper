# contexthelper
[![Go Reference](https://pkg.go.dev/badge/github.com/solsw/contexthelper.svg)](https://pkg.go.dev/github.com/solsw/contexthelper)
[![GitHub](https://img.shields.io/badge/github--green?logo=github)](https://github.com/solsw/contexthelper)

Helpers for Go's [context](https://pkg.go.dev/context) package.

## Installation

```sh
go get github.com/solsw/contexthelper
```

```go
import "github.com/solsw/contexthelper"
```

## Overview

`contexthelper` provides:

- [`Value`](#value) — type-safe retrieval of context values via generics.
- [`AndContext`](#andcontext) — combine two contexts with **AND** semantics.
- [`OrContext`](#orcontext) — combine two contexts with **OR** semantics.

Both `AndContext` and `OrContext` implement the [`context.Context`](https://pkg.go.dev/context#Context) interface.

## Value

```go
func Value[T any](ctx context.Context, key any) (T, bool)
```

`Value` returns the value of type `T` associated with `ctx` for `key`, and a
`bool` reporting whether the value exists and is of type `T`. It is a generic,
type-safe wrapper around [`context.Context.Value`](https://pkg.go.dev/context#Context.Value)
that avoids a manual type assertion at the call site.

```go
type userKey struct{}

ctx := context.WithValue(context.Background(), userKey{}, "alice")

user, ok := contexthelper.Value[string](ctx, userKey{})
// user == "alice", ok == true

n, ok := contexthelper.Value[int](ctx, userKey{})
// n == 0, ok == false  (value present but wrong type)
```

## Combining contexts

Both constructors take two parent contexts. When both parents have a non-nil
`Done()` channel, `AndContext` waits for both channels to close and `OrContext`
waits for either one. If only one parent has a non-nil channel, the combined
context returns that channel directly, without starting a forwarding goroutine.
If both channels are nil, the combined context's `Done()` and `Err()` both return nil.

| Parent deadlines | `AndContext.Deadline()` | `OrContext.Deadline()` |
| --- | --- | --- |
| Both parents have deadlines | Later deadline | Earlier deadline |
| One has a deadline and the other's `Done()` is nil (e.g. `context.Background()`) | Sole deadline | Sole deadline |
| One has a deadline and the other is cancelable with no deadline | No deadline | Sole deadline |
| Neither has a deadline | No deadline | No deadline |

For example, a parent created with `context.WithCancel(context.Background())` can remain active after
the other parent's timeout expires. In that case `AndContext` has no deadline,
so a child created with `context.WithTimeout(andCtx, d)` can enforce its own timeout.

`Err()` returns nil until the combined `Done()` channel closes. The first call
after closure returns the sole non-nil parent error directly or joins both
parent errors if both are present. Later calls return the same error, even if a
second parent has since been canceled. Use `errors.Is` to check for
`context.Canceled` or `context.DeadlineExceeded` in either case.

`Value(key)` returns nil when neither parent has a value for the key. Otherwise
it returns a [`generichelper.Tuple2[any, any]`](https://pkg.go.dev/github.com/solsw/generichelper#Tuple2).
`Item1` holds the first parent's value and `Item2` the second parent's value.
A missing value appears as nil in its parent's position.

## AndContext

```go
func NewAndContext(ctx1, ctx2 context.Context) *AndContext
```

```go
ctx1, cancel1 := context.WithCancel(context.Background())
ctx2, cancel2 := context.WithCancel(context.Background())
defer cancel1()
defer cancel2()

andCtx := contexthelper.NewAndContext(ctx1, ctx2)

cancel1()
fmt.Println(andCtx.Err()) // <nil> because ctx2 is still active

cancel2()
<-andCtx.Done()
fmt.Println(errors.Is(andCtx.Err(), context.Canceled)) // true
```

## OrContext

```go
func NewOrContext(ctx1, ctx2 context.Context) *OrContext
```

```go
ctx1, cancel1 := context.WithCancel(context.Background())
ctx2, cancel2 := context.WithCancel(context.Background())
defer cancel1()
defer cancel2()

orCtx := contexthelper.NewOrContext(ctx1, ctx2)

cancel1()
<-orCtx.Done()
fmt.Println(orCtx.Err() == context.Canceled) // true because only ctx1 is done
```
