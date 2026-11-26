package contexthelper

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/solsw/generichelper"
)

func TestAndContext_Deadline(t *testing.T) {
	bg := context.Background()
	early, late := time.Unix(1, 0), time.Unix(2, 0)
	first, cancelFirst := context.WithDeadline(bg, early)
	defer cancelFirst()
	second, cancelSecond := context.WithDeadline(bg, late)
	defer cancelSecond()
	live, cancelLive := context.WithCancel(bg)
	defer cancelLive()

	tests := []struct {
		name        string
		left, right context.Context
		want        time.Time
		ok          bool
	}{
		{"neither", bg, bg, time.Time{}, false},
		{"first only", first, bg, early, true},
		{"second only", bg, second, late, true},
		{"earlier first", first, second, late, true},
		{"earlier second", second, first, late, true},
		{"unbounded first", live, second, time.Time{}, false},
		{"unbounded second", first, live, time.Time{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NewAndContext(tt.left, tt.right).Deadline()
			if !got.Equal(tt.want) || ok != tt.ok {
				t.Errorf("Deadline() = (%v, %t), want (%v, %t)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestAndContext_ChildDeadline(t *testing.T) {
	bg := context.Background()
	expired, cancelExpired := context.WithDeadline(bg, time.Unix(1, 0))
	defer cancelExpired()
	live, cancelLive := context.WithCancel(bg)
	defer cancelLive()

	parent := NewAndContext(expired, live)
	child, cancelChild := context.WithDeadline(parent, time.Unix(2, 0))
	defer cancelChild()
	if err := child.Err(); err != context.DeadlineExceeded {
		t.Errorf("child.Err() = %v, want context.DeadlineExceeded", err)
	}
}

func TestAndContext_Err(t *testing.T) {
	tests := []struct {
		name       string
		c          *AndContext
		wantErrMsg string
	}{
		{name: "0",
			c:          NewAndContext(context.Background(), context.Background()),
			wantErrMsg: "",
		},
		{name: "1",
			c:          NewAndContext(context.Background(), ctxWithCancel(250*time.Millisecond)),
			wantErrMsg: context.Canceled.Error(),
		},
		{name: "2",
			c:          NewAndContext(ctxWithTimeout(500*time.Millisecond), context.Background()),
			wantErrMsg: context.DeadlineExceeded.Error(),
		},
		{name: "3",
			c:          NewAndContext(ctxWithTimeout(500*time.Millisecond), ctxWithCancel(250*time.Millisecond)),
			wantErrMsg: context.DeadlineExceeded.Error() + "\n" + context.Canceled.Error(),
		},
		{name: "4",
			c:          NewAndContext(ctxWithCancel(250*time.Millisecond), ctxWithTimeout(500*time.Millisecond)),
			wantErrMsg: context.Canceled.Error() + "\n" + context.DeadlineExceeded.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c.Done() == nil {
				if len(tt.wantErrMsg) > 0 {
					t.Errorf("errMsg0 '', wantErrMsg '%s'", tt.wantErrMsg)
				}
				return
			}
			<-tt.c.Done()
			if errMsg1 := tt.c.Err().Error(); errMsg1 != tt.wantErrMsg {
				t.Errorf("errMsg1 '%s', wantErrMsg '%s'", errMsg1, tt.wantErrMsg)
			}
			if errMsg2 := tt.c.Err().Error(); errMsg2 != tt.wantErrMsg {
				t.Errorf("errMsg2 '%s', wantErrMsg '%s'", errMsg2, tt.wantErrMsg)
			}
		})
	}
}

func TestAndContext_Value(t *testing.T) {
	type args struct {
		key any
	}
	tests := []struct {
		name string
		c    *AndContext
		args args
		want any
	}{
		{name: "01",
			c:    NewAndContext(context.Background(), context.Background()),
			args: args{key: "key"},
			want: nil,
		},
		{name: "02",
			c:    NewAndContext(context.WithValue(context.Background(), 1234, 1234), context.WithValue(context.Background(), "1234", "1234")),
			args: args{key: "key"},
			want: nil,
		},
		{name: "1",
			c:    NewAndContext(context.WithValue(context.Background(), "key", 1234), context.Background()),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: nil},
		},
		{name: "2",
			c:    NewAndContext(context.Background(), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: nil, Item2: "1234"},
		},
		{name: "3",
			c:    NewAndContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: "1234"},
		},
		{name: "4",
			c:    NewAndContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "qwerty", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Value(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AndContext.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
