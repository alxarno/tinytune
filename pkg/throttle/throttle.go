package throttle

import (
	"context"
	"time"
)

type Throttler interface {
	Throttle(ctx context.Context) error
}

type throttler struct {
	maxMemoryOccupied float64
	maxWaitingTime    time.Duration
}

//nolint:ireturn
func NewThrottler(maxMemoryOccupied float64, maxWaitingTime time.Duration) Throttler {
	return &throttler{maxMemoryOccupied, maxWaitingTime}
}
