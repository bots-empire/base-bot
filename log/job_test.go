package log

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewJobCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger, obs := NewTest(t, zap.DebugLevel)

	ch := make(chan int, 10)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		tick := time.NewTicker(time.Millisecond * 5)

		var i int
	loop:
		for {
			select {
			case <-tick.C:
				ch <- i
				i++
			case <-ctx.Done():
				close(ch)
				tick.Stop()
				break loop
			}
		}
	}(ctx)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(ctx context.Context, num int) {
			ctx = WithFields(ctx, zap.Int("worker_num", num))
			defer wg.Done()
		loopWorker:
			for {
				localCtx := NewJobCtx(ctx)

				testLocalCtx := NewJobCtx(localCtx)
				assert.Equal(t, JobIDFrom(localCtx), JobIDFrom(testLocalCtx))

				select {
				case j, ok := <-ch:
					if !ok {
						break loopWorker
					}
					Patch(localCtx, logger).Info("tick", zap.Int("iter", j))
					time.Sleep(time.Millisecond * 10)
				case <-ctx.Done():
					break loopWorker
				}
			}
		}(ctx, i)
	}

	time.Sleep(time.Second)
	cancel()
	wg.Wait()

	assert.Greater(t, obs.Len(), 100)
	jobIds := make(map[string]struct{}, obs.Len())
	for _, l := range obs.All() {
		assert.Len(t, l.Context, 3)
		assert.Equal(t, l.Message, "tick")
		for _, c := range l.Context {
			if c.Key == "job_id" {
				jobIds[c.String] = struct{}{}
			}
		}
	}
	assert.Equal(t, len(jobIds), obs.Len())
}
