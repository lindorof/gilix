package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lindorof/gilix"
)

func fdone(t *testing.T) func(ctxRet gilix.RET) {
	return func(ctxRet gilix.RET) {
		t.Logf("!!![%d]", ctxRet)
	}
}
func fworker(t *testing.T, workSleep int, workRet []int) func() TickWorkerMode {
	i := -1
	return func() TickWorkerMode {
		time.Sleep(time.Duration(workSleep) * time.Millisecond)
		i++
		if i >= len(workRet) {
			t.Logf("Quit")
			return TickWorkerModeQuit
		}
		if workRet[i] == 0 {
			t.Logf("NextTick-0")
			return TickWorkerModeNextTick
		}
		t.Logf("Continue-1")
		return TickWorkerModeContinue
	}
}

func TestTickWorker(t *testing.T) {
	cases := []struct {
		ctxTime   int
		tickTime  int
		workSleep int
		workRet   []int
	}{
		{800, 1000, 500, []int{0, 0, 0, 0, 0}},
		{800, 1000, 500, []int{1, 0}},
		{800, 1000, 500, []int{1, 1}},
		{4000, 1000, 500, []int{0, 0, 0, 0}},
		{4000, 1000, 500, []int{1, 0, 1, 0}},
		{4000, 1000, 500, []int{1, 1, 1, 1}},
		{8000, 1000, 500, []int{0, 0, 0, 0}},
		{8000, 1000, 500, []int{1, 0, 1, 0}},
		{8000, 1000, 500, []int{1, 1, 1, 1}},
		{1000, 2000, 500, []int{0, 0}},
		{1000, 2000, 500, []int{1, 0}},
		{1000, 2000, 500, []int{1, 1}},
		{8000, 3000, 500, []int{0, 0, 0, 0}},
		{8000, 3000, 500, []int{1, 0, 1, 0}},
		{8000, 3000, 500, []int{1, 1, 1, 1}},
	}

	for _, c := range cases {
		n := fmt.Sprintf("ctxTime<%d>_tickTime<%d>_workSleep<%d>_workRet%v", c.ctxTime, c.tickTime, c.workSleep, c.workRet)
		t.Run(n, func(t *testing.T) {
			ctx, ctxc := context.WithTimeout(context.Background(), time.Duration(c.ctxTime)*time.Millisecond)
			TickWorker(ctx, time.Duration(c.tickTime)*time.Millisecond, fworker(t, c.workSleep, c.workRet), fdone(t))
			ctxc()
		})
	}
}
