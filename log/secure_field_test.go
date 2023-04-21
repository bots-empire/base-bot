package log

import (
	"reflect"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestSecureWithDebug(t *testing.T) {
	t.Helper()
	l, _ := NewTest(t, zapcore.DebugLevel)
	//l := New(zapcore.DebugLevel)
	type args struct {
		key string
		val string
	}
	tests := []struct {
		name   string
		args   args
		logger Logger
		want   zapcore.Field
	}{
		{
			name: "show_sensitive_data",
			args: args{
				key: "pass",
				val: "super_strong_password",
			},
			logger: l,
			want: Field{
				Key:    "pass",
				Type:   zapcore.StringType,
				String: "super_strong_password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecureField(tt.args.key, tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureField() = %v, want %v", got, tt.want)
			}
			tt.logger.Info("Sensitive data", SecureField(tt.args.key, tt.args.val))
		})
	}
}

func TestSecureWithInfo(t *testing.T) {
	t.Helper()
	l, _ := NewTest(t, zapcore.InfoLevel)

	type args struct {
		key string
		val string
	}
	tests := []struct {
		name   string
		args   args
		logger Logger
		want   zapcore.Field
	}{
		{
			name: "hide_sensitive_data",
			args: args{
				key: "pass",
				val: "super_strong_password",
			},
			logger: l,
			want: Field{
				Key:    "pass",
				Type:   zapcore.StringType,
				String: "***",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SecureField(tt.args.key, tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureField() = %v, want %v", got, tt.want)
			}
			tt.logger.Info("Sensitive data", SecureField(tt.args.key, tt.args.val))
		})
	}
}
