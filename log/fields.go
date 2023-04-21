package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Field = zap.Field

type ctxFieldsKey struct{}

type ctxFields struct {
	mu     *sync.RWMutex
	fields []Field
}

func NewContextWithFields(ctx context.Context) context.Context {
	fields := newCtxFields()
	return context.WithValue(ctx, ctxFieldsKey{}, fields)
}

func newCtxFields() *ctxFields {
	return &ctxFields{
		mu:     &sync.RWMutex{},
		fields: make([]Field, 0),
	}
}

// extract takes the call-scoped ctxFields from context.Context.
// It always returns a non-nil Logger.
func extract(ctx context.Context) *ctxFields {
	l, ok := ctx.Value(ctxFieldsKey{}).(*ctxFields)
	if !ok || l == nil {
		return newCtxFields()
	}
	return l
}

// addFields adds zap fields to the logger.
func addFields(ctx context.Context, fields ...Field) context.Context {
	l := extract(ctx)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append(l.fields, fields...)
	return context.WithValue(ctx, ctxFieldsKey{}, l)
}

// addUniqFields adds zap fields to the logger.
func addUniqFields(ctx context.Context, fields ...Field) context.Context {
	l := extract(ctx)

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, f := range fields {
		exist := false
		for i, logF := range l.fields {
			if logF.Key == f.Key {
				l.fields[i] = f
				exist = true
				break
			}
		}
		if !exist {
			l.fields = append(l.fields, f)
		}
	}

	return context.WithValue(ctx, ctxFieldsKey{}, l)
}
