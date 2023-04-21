package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWithCtx(t *testing.T) {
	logger, obs := NewTest(t, zapcore.DebugLevel)

	type args struct {
		ctx    context.Context
		logger Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "without_any_fields",
			args: args{
				ctx:    context.Background(),
				logger: logger,
			},
		},
		{
			name: "with_fields",
			args: args{
				ctx:    WithFields(context.Background(), zap.String("key", "value")),
				logger: logger,
			},
		},
	}
	msg := "Patch func"
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Patch(tt.args.ctx, tt.args.logger)
			l.Info(msg)
			require.NotNil(t, l)
			require.Equal(t, obs.FilterMessage(msg).All()[i].Message, msg)
			if len(obs.FilterMessage(msg).All()[i].Context) > 0 {
				require.True(t, obs.FilterMessage(msg).All()[i].Context[0].Equals(zap.String("key", "value")))
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	wg := &sync.WaitGroup{}

	logger := New(zap.DebugLevel)
	ctx := WithFields(context.Background(), zap.String("a", "aa"), zap.String("b", "bb"), zap.String("c", "cc"))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			WithFields(ctx, zap.Int(fmt.Sprintf("key=%d", i), i))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			Patch(ctx, logger)
		}
	}()

	wg.Wait()
}

func benchLogger(b *testing.B) Logger {
	b.Helper()

	l := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevel(),
	))

	return &logger{zap: l}
}

// BenchmarkLogger-FromFields  	  148261	      6939 ns/op	      48 B/op	       3 allocs/op
// BenchmarkLogger-FromCtx   	  137953	      7530 ns/op	    1369 B/op	       8 allocs/op
func BenchmarkLoggerFields(b *testing.B) {
	tl := benchLogger(b)
	b.ReportAllocs()
	ctx := WithFields(context.Background(), zap.String("key", "val"), zap.Bool("zap", true))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Patch(ctx, tl).Info("Message ok")
	}
}

// BenchmarkLoggerZap-8      124816       8018 ns/op     1377 B/op        6 allocs/op
func BenchmarkLoggerZap(b *testing.B) {
	b.Helper()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevel(),
	))
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(zap.String("key", "val"), zap.Bool("zap", true)).Info("Message ok")
	}
}

// BenchmarkLoggerStd-8       93962      11416 ns/op     1767 B/op       30 allocs/op
func BenchmarkLoggerStd(b *testing.B) {
	b.Helper()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrus.Fields{"key": "val", "zap": true}).Info("Message ok")
	}
}
