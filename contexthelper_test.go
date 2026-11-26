package contexthelper

import (
	"context"
	"testing"
	"time"
)

func TestSingleCancelableParent(t *testing.T) {
	bg := context.Background()
	canceled, cancel := context.WithCancel(bg)
	cancel()
	expired, release := context.WithDeadline(bg, time.Unix(1, 0))
	defer release()

	for _, parent := range []context.Context{canceled, expired} {
		for name, combined := range map[string]context.Context{
			"and/left":  NewAndContext(parent, bg),
			"and/right": NewAndContext(bg, parent),
			"or/left":   NewOrContext(parent, bg),
			"or/right":  NewOrContext(bg, parent),
		} {
			t.Run(name+"/"+parent.Err().Error(), func(t *testing.T) {
				if combined.Done() != parent.Done() {
					t.Error("Done() should return the sole cancelable parent's channel")
				}
				<-combined.Done()
				if err := combined.Err(); err != parent.Err() {
					t.Errorf("Err() = %T %v, want the original %T", err, err, parent.Err())
				}
			})
		}
	}
}

func TestValue_string(t *testing.T) {
	key1 := "key1"
	key2 := "key2"
	type args struct {
		ctx context.Context
		key any
	}
	tests := []struct {
		name   string
		args   args
		want   string
		wantok bool
	}{
		{name: "1",
			args: args{
				ctx: context.WithValue(context.Background(), key1, key1),
				key: key1,
			},
			want:   key1,
			wantok: true,
		},
		{name: "2",
			args: args{
				ctx: context.WithValue(context.Background(), key1, key1),
				key: key2,
			},
			want:   "",
			wantok: false,
		},
		{name: "3",
			args: args{
				ctx: context.WithValue(context.Background(), key1, 1),
				key: key1,
			},
			want:   "",
			wantok: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotok := Value[string](tt.args.ctx, tt.args.key)
			if got != tt.want {
				t.Errorf("Value[string]() got = %v, want %v", got, tt.want)
				return
			}
			if gotok != tt.wantok {
				t.Errorf("Value[string]() gotok = %v, want %v", gotok, tt.wantok)
			}
		})
	}
}
