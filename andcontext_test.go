package contexthelper

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/solsw/generichelper"
)

func TestAndContext_Deadline(t *testing.T) {
	tests := []struct {
		name string
		c    *AndContext
		want time.Time
		ok   bool
	}{
		{name: "1",
			c:  NewAndContext(context.Background(), context.Background()),
			ok: false,
		},
		{name: "2",
			c:    NewAndContext(ctxWithTimeout(500*time.Millisecond), context.Background()),
			want: time.Now().Add(500 * time.Millisecond).Round(time.Millisecond),
			ok:   true,
		},
		{name: "3",
			c:    NewAndContext(context.Background(), ctxWithTimeout(250*time.Millisecond)),
			want: time.Now().Add(250 * time.Millisecond).Round(time.Millisecond),
			ok:   true,
		},
		{name: "4",
			c:    NewAndContext(ctxWithTimeout(2*time.Second), ctxWithTimeout(4*time.Second)),
			want: time.Now().Add(4 * time.Second).Round(time.Millisecond),
			ok:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := tt.c.Deadline()
			got = got.Round(time.Millisecond)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("andContext.Deadline() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.ok {
				t.Errorf("andContext.Deadline() gotOk = %v, want %v", gotOk, tt.ok)
			}
		})
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
			want: nil,
		},
		{name: "2",
			c:    NewAndContext(context.Background(), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: nil,
		},
		{name: "3",
			c:    NewAndContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "key", "1234")),
			args: args{key: "key"},
			want: generichelper.Tuple2[any, any]{1234, "1234"},
		},
		{name: "4",
			c:    NewAndContext(context.WithValue(context.Background(), "key", 1234), context.WithValue(context.Background(), "qwerty", "1234")),
			args: args{key: "key"},
			want: nil,
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
