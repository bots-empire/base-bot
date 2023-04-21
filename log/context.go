// Package log provides contextual logging.
package log

import (
	"context"
)

type Logger interface {
	Debug(msg string, kv ...Field)
	Info(msg string, kv ...Field)
	Warn(msg string, kv ...Field)
	Error(msg string, kv ...Field)
	With(kv ...Field) Logger
	Named(s string) Logger
	Panic(msg string, kv ...Field)
	Fatal(msg string, kv ...Field)
	Sync() error
}

// Patch - patch logger with Field fields from ctx.
func Patch(ctx context.Context, l Logger) Logger {
	f := extract(ctx)
	if f == nil {
		return l
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return l.With(f.fields...)
}

// WithFields adds fields to context.
func WithFields(ctx context.Context, fields ...Field) context.Context {
	return addFields(ctx, fields...)
}

// WithUniqFields adds uniq fields to context.
func WithUniqFields(ctx context.Context, fields ...Field) context.Context {
	return addUniqFields(ctx, fields...)
}
