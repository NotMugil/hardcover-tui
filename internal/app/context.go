package app

import (
	"context"
	"time"
)

// makeContext creates a context with a standard timeout.
func makeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
