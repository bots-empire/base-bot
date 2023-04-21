package log

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

type logger struct {
	zap *zap.Logger
}

func NewWithLevel(lvl string) (Logger, error) {
	at, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		return nil, err
	}
	return New(at.Level()), nil
}

func New(ll zapcore.Level) Logger {
	encoderCfg := zap.NewProductionEncoderConfig()

	encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder

	l := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevelAt(ll),
	), zap.AddStacktrace(zap.PanicLevel))
	zap.ReplaceGlobals(l)

	return &logger{
		zap: l,
	}
}

func NewNullLogger() Logger {
	return &logger{
		zap: zap.NewNop(),
	}
}

// NewTest - return new test logger
func NewTest(t testing.TB, lvl zapcore.Level) (Logger, *observer.ObservedLogs) {
	l, o := observer.New(lvl)

	testLog := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return l
	})))
	zap.ReplaceGlobals(testLog)
	return &logger{zap: testLog}, o
}

func (l *logger) Debug(msg string, kv ...Field) {
	l.zap.Debug(msg, kv...)
}

func (l *logger) Error(msg string, kv ...Field) {
	l.zap.Error(msg, kv...)
}

func (l *logger) Info(msg string, kv ...Field) {
	l.zap.Info(msg, kv...)
}

func (l *logger) Warn(msg string, kv ...Field) {
	l.zap.Warn(msg, kv...)
}

func (l *logger) With(kv ...Field) Logger {
	return &logger{
		zap: l.zap.With(kv...),
	}
}

func (l *logger) Named(s string) Logger {
	return &logger{
		zap: l.zap.Named(s),
	}
}

func (l *logger) Panic(msg string, kv ...Field) {
	if ce := l.zap.Check(zapcore.PanicLevel, msg); ce != nil {
		ce.Write(kv...)
	}
}

func (l *logger) Fatal(msg string, kv ...Field) {
	if ce := l.zap.Check(zapcore.FatalLevel, msg); ce != nil {
		ce.Write(kv...)
	}
}

func (l *logger) Sync() error {
	return l.zap.Sync()
}
