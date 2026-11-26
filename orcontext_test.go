package contexthelper

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/solsw/generichelper"
)

func TestOrContext_Deadline(t *testing.T) {
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
		{"earlier first", first, second, early, true},
		{"earlier second", second, first, early, true},
		{"unbounded first", live, second, late, true},
		{"unbounded second", first, live, early, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NewOrContext(tt.left, tt.right).Deadline()
			if !got.Equal(tt.want) || ok != tt.ok {
				t.Errorf("Deadline() = (%v, %t), want (%v, %t)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestOrContext_Err(t *testing.T) {
	tests := []struct {
		name       string
		c          *OrContext
		wantErrMsg string
	}{
		{name: "0",
			c:          NewOrContext(context.Background(), context.Background()),
			wantErrMsg: "",
		},
		{name: "2",
			c:          NewOrContext(ctxWithTimeout(500*time.Millisecond), context.Background()),
			wantErrMsg: context.DeadlineExceeded.Error(),
		},
		{name: "3",
			c:          NewOrContext(ctxWithTimeout(10*time.Second), ctxWithCancel(250*time.Millisecond)),
			wantErrMsg: context.Canceled.Error(),
		},
		{name: "4",
			c:          NewOrContext(ctxWithCancel(250*time.Millisecond), ctxWithTimeout(10*time.Second)),
			wantErrMsg: context.Canceled.Error(),
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

func TestOrContext_Value(t *testing.T) {
	type args struct {
		key any
	}
	tests := []struct {
		name string
		c    *OrContext
		args args
		want any
	}{
		{name: "01",
			c:    NewOrContext(context.Background(), context.Background()),
			args: args{key: "key"},
			want: nil,
		},
		{name: "02",
			c:    NewOrContext(context.WithValue(context.Background(), 1234, 1234), context.WithValue(context.Background(), "1234", "1234")),
			args: args{key: "key"},
			want: nil,
		},
		{name: "1",
			c:    NewOrContext(context.WithValue(context.Background(), "key", 1234), context.Background()),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: nil},
		},
		{name: "2",
			c:    NewOrContext(context.Background(), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: nil, Item2: "1234"},
		},
		{name: "3",
			c:    NewOrContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: "1234"},
		},
		{name: "4",
			c:    NewOrContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "qwerty", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{Item1: 1234, Item2: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Value(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrContext.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
