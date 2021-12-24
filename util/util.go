package util

import (
	"context"
	"strings"
	"time"

	"github.com/lindorof/gilix"
)

func ContextErr(ctx context.Context) gilix.RET {
	if ctx == nil {
		return gilix.RET_SUCCESS
	}

	err := ctx.Err()
	if err == nil {
		return gilix.RET_SUCCESS
	}
	if strings.Contains(strings.ToLower(err.Error()), "deadline") {
		return gilix.RET_TIMEOUT
	}
	if strings.Contains(strings.ToLower(err.Error()), "cancel") {
		return gilix.RET_CANCELLED
	}

	return gilix.RET_CANCELLED
}

func TickWorker(ctx context.Context, d time.Duration, fworker func() (quit bool), fdone func(ctxErr gilix.RET)) {
	ticker := time.NewTicker(d)
LOOP:
	for {
		select {
		case <-ctx.Done():
			fdone(ContextErr(ctx))
			break LOOP
		case <-ticker.C:
			if fworker() {
				break LOOP
			}
		}
	}
	ticker.Stop()
}
