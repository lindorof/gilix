package util

import (
	"context"
	"time"

	"github.com/lindorof/gilix"
)

type TickWorkerMode int

const (
	TickWorkerModeQuit     = 1
	TickWorkerModeNextTick = 2
	TickWorkerModeContinue = 3
)

func TickWorker(ctx context.Context, d time.Duration, fworker func() TickWorkerMode, fdone func(ctxRet gilix.RET)) {
	ticker := time.NewTicker(d)
LOOP:
	for {
		select {
		case <-ctx.Done():
			fdone(ContextRet(ctx))
			break LOOP
		case <-ticker.C:
			for {
				m := fworker()
				if m == TickWorkerModeQuit {
					break LOOP
				}
				if m == TickWorkerModeNextTick {
					break
				}
				if m == TickWorkerModeContinue {
					continue
				}
			}
		}
	}
	ticker.Stop()
}
