package setup

import (
	"context"
	"time"
)

func makeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
