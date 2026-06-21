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

## AndContext

```go
func NewAndContext(ctx1, ctx2 context.Context) *AndContext
```

`AndContext` combines two contexts with **AND** semantics: it is considered
done only when *both* combined contexts are done.

- **Done** — the channel returned by `Done` is closed when all non-nil `Done`
  channels of the combined contexts are closed. A context with a nil `Done`
  channel (e.g. [`context.Background`](https://pkg.go.dev/context#Background))
  is treated as already done.
- **Deadline** — if both contexts have a deadline, the **latest** one is returned.
- **Err** — `nil` until `Done` is closed; afterwards it returns an error that
  [joins](https://pkg.go.dev/errors#Join) both contexts' `Err`s.
- **Value** — returns a [`generichelper.Tuple2[any, any]`](https://pkg.go.dev/github.com/solsw/generichelper#Tuple2)
  holding each combined context's value for the key (an item is `nil` if that
  context does not hold the key), or `nil` if neither context holds the key.

```go
ctx1, cancel1 := context.WithCancel(context.Background())
ctx2, cancel2 := context.WithCancel(context.Background())
defer cancel1()
defer cancel2()

andCtx := contexthelper.NewAndContext(ctx1, ctx2)

cancel1()
// andCtx is NOT done yet — ctx2 is still active.

cancel2()
<-andCtx.Done() // now closed — both contexts are done.
```

## OrContext

```go
func NewOrContext(ctx1, ctx2 context.Context) *OrContext
```

`OrContext` combines two contexts with **OR** semantics: it is considered done
as soon as *either* combined context is done.

- **Done** — the channel returned by `Done` is closed when either context's
  `Done` channel is closed.
- **Deadline** — if both contexts have a deadline, the **earliest** one is returned.
- **Err** — `nil` until `Done` is closed; afterwards it returns the non-nil
  `Err`(s) of the combined contexts, [joined](https://pkg.go.dev/errors#Join)
  when both are done.
- **Value** — returns a [`generichelper.Tuple2[any, any]`](https://pkg.go.dev/github.com/solsw/generichelper#Tuple2)
  holding each combined context's value for the key (an item is `nil` if that
  context does not hold the key), or `nil` if neither context holds the key.

```go
ctx1, cancel1 := context.WithCancel(context.Background())
ctx2, cancel2 := context.WithCancel(context.Background())
defer cancel1()
defer cancel2()

orCtx := contexthelper.NewOrContext(ctx1, ctx2)

cancel1()
<-orCtx.Done() // closed as soon as ctx1 is done.
```
