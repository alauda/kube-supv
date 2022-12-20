package utils

import (
	"context"
	"time"
)

const DefaultTimeoutSeconds = 10

func DefaultTimeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*time.Duration(DefaultTimeoutSeconds))
}
