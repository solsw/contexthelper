package contexthelper

import (
	"context"
	"testing"
)

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
