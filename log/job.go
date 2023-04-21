package log

import (
	"context"
	"math/rand"

	"go.uber.org/zap"
)

type (
	keyJobID struct{}
)

func WithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, keyJobID{}, jobID)
}

func JobIDFrom(ctx context.Context) string {
	v, ok := ctx.Value(keyJobID{}).(string)
	if ok {
		return v
	}
	return "not found"
}

const randBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// NewJobID returns new pseudo-random job ID string.
func NewJobID() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = randBytes[rand.Int63()%int64(len(randBytes))] // #nosec
	}
	return string(b)
}

func isJobIDExist(ctx context.Context) bool {
	_, ok := ctx.Value(keyJobID{}).(string)
	return ok
}

func NewJobCtx(ctx context.Context) context.Context {
	// If exist in ctx then not need generate new job id
	if isJobIDExist(ctx) {
		return ctx
	}

	jobID := NewJobID()
	ctx = WithJobID(ctx, jobID)

	return WithUniqFields(ctx, zap.String("job_id", jobID))
}
